package helpers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/auth"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/gits"
	"golang.org/x/oauth2"

	"github.com/cenkalti/backoff"
	"github.com/jenkins-x/bdd-jx/test/utils/parsers"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"

	"github.com/jenkins-x/bdd-jx/test/utils"

	"github.com/jenkins-x/bdd-jx/test/utils/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	scm "github.com/jenkins-x/go-scm/scm"
	scmFactory "github.com/jenkins-x/go-scm/scm/factory"
)

const (
	// BDDPullRequestApproverUsernameEnvVar is the environment variable that we look at for the username for the approver in some tests
	BDDPullRequestApproverUsernameEnvVar = "BDD_APPROVER_USERNAME"
	// BDDPullRequestApproverTokenEnvVar is the environment variable that we look at for the token for the approver in some tests.
	BDDPullRequestApproverTokenEnvVar = "BDD_APPROVER_ACCESS_TOKEN"
)

var (
	// TempDirPrefix The prefix to append to applicationss created in testing
	TempDirPrefix = "bdd-"
	// WorkDir The current working directory
	WorkDir              string
	DefaultRepositoryURL = "http://chartmuseum.jenkins-x.io"

	// all timeout values are in minutes
	// timeout for a build to complete successfully
	TimeoutBuildCompletes = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_COMPLETES", 40)

	// TimeoutBuildIsRunningInStaging Timeout for promoting an application to staging environment
	TimeoutBuildIsRunningInStaging = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING", 20)

	// TimeoutPipelineActivityComplete for promoting an application to staging environment
	TimeoutPipelineActivityComplete = utils.GetTimeoutFromEnv("BDD_TIMEOUT_PIPELINE_ACTIVITY_COMPLETE", 15)

	// TimeoutUrlReturns Timeout for a given URL to return an expected status code
	TimeoutUrlReturns = utils.GetTimeoutFromEnv("BDD_TIMEOUT_URL_RETURNS", 15)

	// TimeoutPreviewUrlReturns Timeout for a preview URL to be available
	TimeoutPreviewUrlReturns = utils.GetTimeoutFromEnv("BDD_TIMEOUT_PREVIEW_URL_RETURNS", 15)

	// TimeoutCmdLine Timeout to wait for a command line execution to complete
	TimeoutCmdLine = utils.GetTimeoutFromEnv("BDD_TIMEOUT_CMD_LINE", 1)

	// TimeoutSessionWait Session wait timeout
	TimeoutSessionWait = utils.GetTimeoutFromEnv("BDD_TIMEOUT_SESSION_WAIT", 60)

	// TimeoutDeploymentRollout defines the timeout waiting for a deployment rollout
	TimeoutDeploymentRollout = utils.GetTimeoutFromEnv("", 3)

	// InsecureURLSkipVerify skips the TLS verify when checking URLs of deployed applications
	InsecureURLSkipVerify = utils.GetEnv("BDD_URL_INSECURE_SKIP_VERIFY", "false")
	// TimeoutProwActionWait defines the timeout for waiting for a prow action to complete
	TimeoutProwActionWait = utils.GetTimeoutFromEnv("BDD_TIMEOUT_PROW_ACTION_WAIT", 5)

	// EnableChatOpsTests turns on the chatops tests when specified as true
	EnableChatOpsTests = utils.GetEnv("BDD_ENABLE_TEST_CHATOPS_COMMANDS", "false")

	// DisablePipelineActivityCheck turns off the check for updated PipelineActivity. Meant to be used with static masters.
	DisablePipelineActivityCheck = utils.GetEnv("BDD_DISABLE_PIPELINEACTIVITY_CHECK", "false")

	// PullRequestApproverUsername is the username used for /approve commands on PRs, since the bot user may not be able to.
	PullRequestApproverUsername = utils.GetEnv(BDDPullRequestApproverUsernameEnvVar, "")

	// PullRequestApproverToken is the access token used by the PullRequestApproverUsername user.
	PullRequestApproverToken = utils.GetEnv(BDDPullRequestApproverTokenEnvVar, "")
)

// TestOptions is the base testing object
type TestOptions struct {
	Interactive     bool
	WorkDir         string
	ApplicationName string
	Organisation    string
}

func AssignWorkDirValue(generatedWorkDir string) {
	WorkDir = generatedWorkDir
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func (t *TestOptions) GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = l.Close()
	}()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// GetGitOrganisation Gets the current git organisation/user
func (t *TestOptions) GetGitOrganisation() string {
	org := os.Getenv("GIT_ORGANISATION")
	return org
}

// GetGitProvider returns a git provider that uses default credentials stored in the jx-auth-configmap or in ~/.jx/gitAuth.yaml
func (t *TestOptions) GetGitProvider() (gits.GitProvider, error) {
	return t.getGitProviderWithUserFunc(func(service auth.ConfigService, config *auth.AuthConfig, server *auth.AuthServer) (*auth.UserAuth, error) {
		return config.CurrentUser(server, false), nil
	})
}

// GetApproverGitProvider returns a git provider that uses credentials for the approver user defined in environment variables
// We don't use standard user auth here because the user/token isn't defined during boot, so we'll create the credentials
// for the user on the fly.
func (t *TestOptions) GetApproverGitProvider() (gits.GitProvider, error) {
	return t.getGitProviderWithUserFunc(func(service auth.ConfigService, config *auth.AuthConfig, server *auth.AuthServer) (*auth.UserAuth, error) {
		userAuth := config.FindUserAuth(server.URL, PullRequestApproverUsername)
		if userAuth == nil {
			userAuth = config.GetOrCreateUserAuth(server.URL, PullRequestApproverUsername)
			userAuth.ApiToken = PullRequestApproverToken
			userAuth.Password = PullRequestApproverToken
			err := service.SaveConfig()
			if err != nil {
				return nil, err
			}
		}
		return userAuth, nil
	})
}

// getGitProviderWithUserFunc returns a git provider that uses default credentials stored in the jx-auth-configmap or in ~/.jx/gitAuth.yaml
func (t *TestOptions) getGitProviderWithUserFunc(userAuthFunc func(auth.ConfigService, *auth.AuthConfig, *auth.AuthServer) (*auth.UserAuth, error)) (gits.GitProvider, error) {
	factory := cmd.NewFactory()
	_, ns, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	authConfigService, err := factory.CreateGitAuthConfigService(ns, "")
	if err != nil {
		return nil, err
	}

	config, err := authConfigService.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading auth config: %s", err)
	}

	if config == nil {
		return nil, fmt.Errorf("auth config is nil but no error was returned by LoadConfig")
	}

	authServer := config.CurrentAuthServer()
	if authServer == nil {
		return nil, fmt.Errorf("no config for git auth server found")
	}
	userAuth, err := userAuthFunc(authConfigService, config, authServer)
	if err != nil {
		return nil, err
	}
	if userAuth == nil {
		return nil, fmt.Errorf("no config for git user auth found")
	}

	gitProvider, err := gits.CreateProvider(authServer, userAuth, nil)
	if err != nil {
		return nil, err
	}
	return gitProvider, nil
}

// GitProviderURL Gets the current git provider URL
func (t *TestOptions) GitProviderURL() (string, error) {
	gitProviderURL := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderURL != "" {
		return gitProviderURL, nil
	}
	var out string
	By("running jx get gitserver", func() {

		r := runner.New(t.WorkDir, nil, 0)
		var err error
		out, err = r.RunWithOutput("get", "gitserver")
		utils.ExpectNoError(err)
	})
	var gitServers []parsers.GitServer
	var err error
	By("parsing the output of jx get gitserver", func() {
		gitServers, err = parsers.ParseJxGetGitServer(out)
	})
	if err != nil {
		return "", err
	}
	if len(gitServers) < 1 {
		return "", errors.Errorf("Must be at least 1 git server configured")
	}

	return gitServers[0].Url, nil
}

func (t *TestOptions) GitHubClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: t.GitHubToken()},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	Expect(client).ShouldNot(BeNil())
	return client
}

// GitHubToken returns the GitHub token for the pipeline user.
func (t *TestOptions) GitHubToken() string {
	provider, err := t.GetGitProvider()
	Expect(err).Should(BeNil())

	return provider.UserAuth().ApiToken
}

// GitOpsDevRepo returns repository URL for the gitops environment repo.
// The empty string is returned in case there is no gitops repo.
func (t *TestOptions) GitOpsDevRepo() string {
	args := []string{"get", "environment", "dev", "-o=jsonpath='{.spec.source.url}'"}
	command := exec.Command("kubectl", args...)
	session, err := gexec.Start(command, nil, nil)
	Expect(err).Should(BeNil())

	session.Wait(TimeoutCmdLine)
	Eventually(session).Should(gexec.Exit(0))

	url := strings.Trim(string(session.Out.Contents()), "'")
	return url
}

// GitOpsEnabled returns true if the current cluster is GitOps enabled, false otherwise.
func (t *TestOptions) GitOpsEnabled() bool {
	url := t.GitOpsDevRepo()
	if url == "" {
		return false
	} else {
		return true
	}
}

// NextBuildNumber returns the next build number for a given repo by looking at the SourceRepository CRD.
func (t *TestOptions) NextBuildNumber(repo *gits.GitRepository) string {
	crd := strings.ToLower(fmt.Sprintf("%s-%s", repo.Organisation, repo.Name))

	args := []string{"get", "sourcerepository", crd, "-o", "json"}
	command := exec.Command("kubectl", args...)
	session, err := gexec.Start(command, nil, nil)
	Expect(err).Should(BeNil())

	session.Wait(TimeoutCmdLine)
	Eventually(session).Should(gexec.Exit(0))

	out := string(session.Out.Contents())
	sourceRepository := v1.SourceRepository{}
	err = json.Unmarshal([]byte(out), &sourceRepository)
	Expect(err).Should(BeNil())

	latestBuild := sourceRepository.Annotations["jenkins.io/last-build-number-for-master"]
	if latestBuild == "" {
		latestBuild = "0"
	}
	latestBuildInt, err := strconv.Atoi(latestBuild)
	Expect(err).Should(BeNil())

	nextBuildInt := latestBuildInt + 1

	return strconv.Itoa(nextBuildInt)
}

// GetPullTitleForBranch returns the PullTitle field from the PipelineActivity for the owner/repo/branch
func (t *TestOptions) GetPullTitleFromActivity(owner string, repo string, branch string, buildNumber int) string {
	activityName := fmt.Sprintf("%s-%s-%s-%s", owner, repo, branch, strconv.Itoa(buildNumber))
	args := []string{"get", "pipelineactivity", activityName, "-o=jsonpath='{.spec.pullTitle}'"}

	command := exec.Command("kubectl", args...)
	session, err := gexec.Start(command, nil, nil)
	Expect(err).Should(BeNil())

	session.Wait(TimeoutCmdLine)
	Eventually(session).Should(gexec.Exit(0))

	pullTitle := strings.Trim(string(session.Out.Contents()), "'")
	return pullTitle
}

func (t *TestOptions) TheApplicationIsRunningInProduction(statusCode int) {
	t.TheApplicationIsRunning(statusCode, "production")
}

// TheApplicationIsRunningInStaging lets assert that the application is deployed into the first automatic staging environment
func (t *TestOptions) TheApplicationIsRunningInStaging(statusCode int) {
	t.TheApplicationIsRunning(statusCode, "staging")
}

// TheApplicationIsRunning lets assert that the application is deployed into the passed environment
func (t *TestOptions) TheApplicationIsRunning(statusCode int, environment string) {
	u := ""
	args := []string{"get", "applications", "-e", environment}
	r := runner.New(t.WorkDir, nil, 0)
	argsStr := strings.Join(args, " ")
	f := func() error {
		var err error
		var out string
		By(fmt.Sprintf("running jx %s", argsStr), func() {
			out, err = r.RunWithOutput(args...)
			utils.ExpectNoError(err)
		})
		var applications map[string]parsers.Application
		By(fmt.Sprintf("parsing the output of jx %s", argsStr), func() {
			applications, err = parsers.ParseJxGetApplications(out)
		})
		if err != nil {
			// Need to do return an error here to perform a retry and backoff
			utils.LogInfof("failed to parse applications: %s\n", err.Error())
			return err
		}

		applicationName := t.GetApplicationName()
		var application *parsers.Application
		By(fmt.Sprintf("validating that the application %s was returned by jx %s", applicationName, argsStr), func() {
			application, err = getApplication(applicationName, applications)
		})
		if err != nil {
			utils.LogInfof("failed to get application: %s. Output of jx %s was %s. Parsed applications map is %v`\n", err.Error(), argsStr, out, applications)
			return err
		}
		Expect(application).ShouldNot(BeNil(), "no application found for % in environment %s", applicationName, environment)
		By(fmt.Sprintf("getting url for application %s", application.Name), func() {
			u = application.Url
		})
		if u == "" {
			return fmt.Errorf("no URL found for environment %s has app: %#v", environment, applications)
		}
		utils.LogInfof("still looking for application %s in env %s\n", applicationName, environment)
		return nil
	}

	By(fmt.Sprintf("retrying jx %s with exponential backoff", argsStr), func() {
		err := RetryExponentialBackoff(TimeoutBuildIsRunningInStaging, f)
		Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")
	})

	By(fmt.Sprintf("getting %s", u), func() {
		Expect(u).ShouldNot(BeEmpty(), "no URL for environment %s", environment)
		err := t.ExpectUrlReturns(u, statusCode, TimeoutUrlReturns)
		Expect(err).ShouldNot(HaveOccurred(), fmt.Sprintf("request application URL should return %d", statusCode))
	})
}

// WaitForDeployment waits for the specified deployment to rollout. Wait timeout can be set via BDD_DEPLOYMENT_ROLLOUT_WAIT.
func (t *TestOptions) WaitForDeploymentRollout(deployment string) {
	args := []string{"rollout", "status", "-w", fmt.Sprintf("deployment/%s", deployment)}
	command := exec.Command("kubectl", args...)
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).Should(BeNil())

	session.Wait(TimeoutDeploymentRollout)
	Eventually(session).Should(gexec.Exit())
}

func getApplication(applicationName string, runningApplications map[string]parsers.Application) (*parsers.Application, error) {
	if len(runningApplications) == 0 {
		return nil, fmt.Errorf("no applications found")
	}

	applicationEnvInfo, ok := runningApplications[applicationName]
	if !ok {
		applicationName = "jx-" + applicationName
		applicationEnvInfo, ok = runningApplications[applicationName]
		if !ok {
			utils.LogInfof("applications found were %v\n", runningApplications)
		}
	}
	return &applicationEnvInfo, nil
}

// TheApplicationShouldBeBuiltAndPromotedViaCICD asserts that the project
// should be created in Jenkins and that the build should complete successfully
func (t *TestOptions) TheApplicationShouldBeBuiltAndPromotedViaCICD(statusCode int) {
	applicationName := t.GetApplicationName()
	owner := t.GetGitOrganisation()
	jobName := owner + "/" + applicationName + "/master"

	By(fmt.Sprintf("checking that job %s completes successfully", jobName), func() {
		t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)
	})
	By("checking that the application is running in staging", func() {
		t.TheApplicationIsRunningInStaging(statusCode)
	})
}

// CreatePullRequestAndGetPreviewEnvironment asserts that a pull request can be created
// on the application and the PR goes green and a preview environment is available
func (t *TestOptions) CreatePullRequestAndGetPreviewEnvironment(statusCode int) error {
	applicationName := t.GetApplicationName()
	workDir := filepath.Join(t.WorkDir, applicationName)
	owner := t.GetGitOrganisation()
	r := runner.New(workDir, nil, 0)

	By(fmt.Sprintf("creating a pull request in directory %s", workDir), func() {
		t.ExpectCommandExecution(workDir, TimeoutCmdLine, 0, "git", "checkout", "-b", "changes")
	})

	By("making a code change, committing and pushing it", func() {
		// now lets make a code change
		fileName := "README.md"
		readme := filepath.Join(workDir, fileName)

		data := []byte("My First PR/n")
		err := ioutil.WriteFile(readme, data, util.DefaultWritePermissions)
		if err != nil {
			panic(err)
		}

		t.ExpectCommandExecution(workDir, time.Minute, 0, "git", "add", fileName)
		t.ExpectCommandExecution(workDir, time.Minute, 0, "git", "commit", "-a", "-m", "My first PR commit")
		t.ExpectCommandExecution(workDir, time.Minute, 0, "git", "push", "--set-upstream", "origin", "changes")
	})

	prTitle := "My First PR commit"
	args := []string{"create", "pullrequest", "-b", "--title", prTitle, "--body", "PR comments"}
	argsStr := strings.Join(args, " ")
	var out string
	By(fmt.Sprintf("creating a pull request by running jx %s", argsStr), func() {
		var err error
		out, err = r.RunWithOutputNoTimeout(args...)
		out = strings.TrimSpace(out)
		if err != nil {
			utils.LogInfof("ERROR: %s\n", err.Error())
		} else {
			Expect(out).ShouldNot(BeEmpty(), "no output returned from command: jx "+argsStr)
		}
		utils.ExpectNoError(err)
	})

	utils.LogInfof("running jx %s and got result: %s\n", argsStr, out)

	var pr *parsers.CreatePullRequest
	var err error
	By(fmt.Sprintf("parsing the output %s of jx %s", out, argsStr), func() {
		pr, err = parsers.ParseJxCreatePullRequest(out)
		utils.ExpectNoError(err)
	})

	var prNumber int
	By(fmt.Sprintf("validating that the pull request %v exists and has a number", pr), func() {
		Expect(pr).ShouldNot(BeNil())
		prNumber = pr.PullRequestNumber
		Expect(prNumber).ShouldNot(BeNil())

	})

	buildNumber := 0
	jobName := owner + "/" + applicationName + "/PR-" + strconv.Itoa(prNumber)
	By(fmt.Sprintf("checking that job %s completes successfully", jobName), func() {
		buildNumber = t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)
		utils.ExpectNoError(err)
	})
	if t.ShouldTestPipelineActivityUpdate() {
		By("verifying that PipelineActivity has been updated to include the pull request title", func() {
			pullTitle := t.GetPullTitleFromActivity(owner, applicationName, "pr-"+strconv.Itoa(prNumber), buildNumber)
			Expect(pullTitle).Should(Equal(prTitle))
		})
	}

	args = []string{"get", "previews"}
	argsStr = strings.Join(args, " ")
	By(fmt.Sprintf("verifying there is a preview environment by running jx %s", argsStr), func() {
		var err error
		out, err = r.RunWithOutput(args...)
		utils.ExpectNoError(err)
	})

	logError := func(err error) error {
		utils.LogInfof("WARNING: %s\n", err.Error())
		return err
	}

	f := func() error {
		var err error
		var previews map[string]parsers.Preview

		utils.LogInfof("parsing the output of jx %s", argsStr)
		out, err = r.RunWithOutput(args...)
		if err != nil {
			return logError(err)
		}
		previews, err = parsers.ParseJxGetPreviews(out)
		if err != nil {
			return logError(err)
		}
		previewEnv := previews[pr.Url]
		applicationUrl := previewEnv.Url
		if applicationUrl == "" {
			idx := strings.LastIndex(pr.Url, "/")
			for k, v := range previews {
				utils.LogInfof("found Preview URL %s with preview %s", k, v.Url)
				if idx > 0 {
					if strings.HasSuffix(k, pr.Url[idx:]) {
						applicationUrl = v.Url
						utils.LogInfof("for PR %s using preview %s", k, applicationUrl)
					}
				}
			}
		}
		if applicationUrl == "" {
			return logError(fmt.Errorf("no Preview Application URL found for PR %s", pr.Url))
		}

		utils.LogInfof("Running Preview Environment application at: %s\n", util.ColorInfo(applicationUrl))

		err = t.ExpectUrlReturns(applicationUrl, statusCode, TimeoutUrlReturns)
		if err != nil {
			return logError(fmt.Errorf("preview URL at %s not working: %s", applicationUrl, err.Error()))
		}
		return nil
	}

	By(fmt.Sprint("retrying waiting for Preview URL to be working with exponential backoff to ensure it completes"), func() {
		err := Retry(TimeoutPreviewUrlReturns, f)
		Expect(err).ShouldNot(HaveOccurred(), "preview environment visible at a URL")
	})
	return nil

}

// SetGitHubToken runs jx create git token using the values of GIT_ORGANISATION & GH_ACCESS_TOKEN
func (t *TestOptions) SetGitHubToken() {
	gitUser, set := os.LookupEnv("GIT_ORGANISATION")
	if !set {
		Fail("GIT_ORGANISATION environment variable must be set")
	}

	token, set := os.LookupEnv("GH_ACCESS_TOKEN")
	if !set {
		Fail("GH_ACCESS_TOKEN environment variable must be set")
	}

	args := []string{"create", "git", "token", gitUser, "-t", token}
	command := exec.Command(runner.JxBin(), args...)
	session, err := gexec.Start(command, nil, nil)
	Expect(err).Should(BeNil())

	session.Wait(TimeoutCmdLine)
	Eventually(session).Should(gexec.Exit(0))
}

// GetPullRequestWithTitle Returns a pull request with a matching title
func (t *TestOptions) GetPullRequestWithTitle(provider gits.GitProvider, repoOwner string, repoName string, title string) (*gits.GitPullRequest, error) {
	pullRequestList, err := provider.ListOpenPullRequests(repoOwner, repoName)
	if err != nil {
		return nil, err
	}

	for _, pullRequest := range pullRequestList {
		if pullRequest.Title == title {
			return pullRequest, nil
		}
	}

	return nil, nil
}

// ApprovePullRequestFromLogOutput takes the default provider, the approver user's provider, git info, and the output from a command that
// created a PR, and adds the approver user as a collaborator, accepts the invitation, and approves the PR.
func (t *TestOptions) ApprovePullRequestFromLogOutput(provider gits.GitProvider, approverProvider gits.GitProvider, gitInfo *gits.GitRepository, output string) {
	createdPR, err := parsers.ParseJxCreatePullRequestFromFullLog(output)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(createdPR).ShouldNot(BeNil())

	repoStruct := &gits.GitRepository{
		Name:         gitInfo.Name,
		Organisation: gitInfo.Organisation,
	}
	pr, err := provider.GetPullRequest(gitInfo.Organisation, repoStruct, createdPR.PullRequestNumber)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(pr).ShouldNot(BeNil())
	Expect(*pr.State).Should(Equal("open"))

	By("adding the approver user as a collaborator")
	err = t.AddApproverAsCollaborator(provider, gitInfo.Organisation, gitInfo.Name)
	Expect(err).ShouldNot(HaveOccurred())

	By("approving the PR")
	err = t.ApprovePullRequest(provider, approverProvider, pr)
	Expect(err).ShouldNot(HaveOccurred())
}

// AddApproverAsCollaborator adds the approver user as a collaborator to the given repo, and accepts the invitation.
func (t *TestOptions) AddApproverAsCollaborator(provider gits.GitProvider, repoOwner string, repoName string) error {
	err := provider.AddCollaborator(PullRequestApproverUsername, repoOwner, repoName)
	if err != nil {
		return err
	}
	approverProvider, err := t.GetApproverGitProvider()
	if err != nil {
		return err
	}
	invites, _, err := approverProvider.ListInvitations()
	if err != nil {
		return err
	}
	for _, x := range invites {
		// Accept all invitations for the pipeline user
		_, err = approverProvider.AcceptInvitation(*x.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPullRequestByNumber Returns a pull request with the given owner, repo, and number
func (t *TestOptions) GetPullRequestByNumber(provider gits.GitProvider, repoOwner string, repoName string, prNumber int) (*gits.GitPullRequest, error) {
	repoStruct := &gits.GitRepository{
		Name:         repoName,
		Organisation: repoOwner,
	}
	pr, err := provider.GetPullRequest(repoOwner, repoStruct, prNumber)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// TailBuildLog tails the logs of the specified job
func (t *TestOptions) TailBuildLog(jobName string, maxDuration time.Duration) {
	args := []string{"get", "build", "logs", "--wait", jobName}
	argsStr := strings.Join(args, " ")
	By(fmt.Sprintf("checking that there is a job built successfully by calling jx %s", argsStr), func() {
		t.ExpectJxExecution(t.WorkDir, maxDuration, 0, args...)
	})
}

// ThereShouldBeAJobThatCompletesSuccessfully asserts that the given job name completes within the given duration
func (t *TestOptions) ThereShouldBeAJobThatCompletesSuccessfully(jobName string, maxDuration time.Duration) int {
	t.TailBuildLog(jobName, maxDuration)

	r := runner.New(t.WorkDir, nil, 0)
	// TODO the current --build 1 breaks as it can be number 2 these days!
	//out := r.RunWithOutput("get", "activities", "--filter", jobName, "--build", "1")
	args := []string{"get", "activities", "--filter", jobName}
	argsStr := strings.Join(args, " ")
	var activities map[string]*parsers.Activity
	f := func() error {
		var err error
		var out string
		By(fmt.Sprintf("calling jx %s", argsStr), func() {
			out, err = r.RunWithOutput(args...)
		})
		out, err = r.RunWithOutput(args...)
		if err != nil {
			return err
		}
		activities, err = parsers.ParseJxGetActivities(out)
		// TODO fails on --ng for now...
		//utils.ExpectNoError(err)
		if err != nil {
			utils.LogInfof("got error parsing activities: %s\n", err.Error())
		}
		return err
	}

	// Sleep 15 seconds to make sure that PipelineActivity gets updated after the run has completed
	time.Sleep(15 * time.Second)
	By(fmt.Sprintf("retrying jx %s with exponential backoff to ensure it completes", argsStr), func() {
		err := RetryExponentialBackoff(TimeoutPipelineActivityComplete, f)
		Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")
	})

	buildNumber := 1
	activityKey := fmt.Sprintf("%s #%d", jobName, 1)
	By(fmt.Sprintf("finding the activity for %s in %v", activityKey, activities), func() {
		if activities != nil {
			Expect(activities).Should(HaveLen(1), fmt.Sprintf("should be one activity but found %d having run jx get activities --filter %s --build 1; activities %v", len(activities), jobName, activities))
			activity, ok := activities[fmt.Sprintf("%s #%d", jobName, 1)]
			if !ok {
				// TODO lets see if the build is number 2 instead which it is for tekton currently
				activity, ok = activities[fmt.Sprintf("%s #%d", jobName, 2)]
				buildNumber = 2
			}
			Expect(ok).Should(BeTrue(), fmt.Sprintf("could not find job with name %s #1 or #2", jobName))
			utils.LogInfof("build status for '%s' is '%s'\n", jobName+"-"+strconv.Itoa(buildNumber), activity.Status)

			// TODO: Fix the regex in get_activities_parser to not treat "Succeeded Version: 0.0.1" as the status. I'm too lazy right now.
			Expect(activity.Status).Should(HavePrefix("Succeeded"))
		}
	})

	return buildNumber
}

// RetryExponentialBackoff retries the given function up to the maximum duration
func Retry(maxDuration time.Duration, f func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxDuration
	exponentialBackOff.MaxInterval = 20 * time.Second
	exponentialBackOff.Reset()
	utils.LogInfof("retrying for duration %#v with max interval %#v\n", maxDuration, exponentialBackOff.MaxInterval)
	err := backoff.Retry(f, exponentialBackOff)
	return err
}

// RetryExponentialBackoff retries the given function up to the maximum duration
func RetryExponentialBackoff(maxDuration time.Duration, f func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxDuration
	exponentialBackOff.Reset()
	err := backoff.Retry(f, exponentialBackOff)
	return err
}

// GetApplicationName gets the application name for the current test case
func (t *TestOptions) GetApplicationName() string {
	applicationName := t.ApplicationName
	if applicationName == "" {
		_, applicationName = filepath.Split(t.WorkDir)
	}
	return applicationName
}

// ExpectCommandExecution performs the given command in the current work directory and asserts that it completes successfully
func (t *TestOptions) ExpectCommandExecution(dir string, commandTimeout time.Duration, exitCode int, c string, args ...string) {
	f := func() error {
		command := exec.Command(c, args...)
		command.Dir = dir
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		session.Wait(commandTimeout)
		Eventually(session).Should(gexec.Exit(exitCode))
		return err
	}
	err := RetryExponentialBackoff(TimeoutCmdLine, f)
	Expect(err).ShouldNot(HaveOccurred())
}

func (t *TestOptions) ExpectJxExecution(dir string, commandTimeout time.Duration, exitCode int, args ...string) {
	r := runner.New(dir, &commandTimeout, exitCode)
	r.Run(args...)
}

func (t *TestOptions) ExpectJxExecutionWithOutput(dir string, commandTimeout time.Duration, exitCode int, args ...string) string {
	r := runner.New(dir, &commandTimeout, exitCode)
	out, err := r.RunWithOutput(args...)
	utils.ExpectNoError(err)
	return out
}

// DeleteApplications should we delete applications after the quickstart has run
func (t *TestOptions) DeleteApplications() bool {
	text := os.Getenv("JX_DISABLE_DELETE_APP")
	return strings.ToLower(text) != "true"
}

// DeleteRepos should we delete the git repos after the quickstart has run
func (t *TestOptions) DeleteRepos() bool {
	text := os.Getenv("JX_DISABLE_DELETE_REPO")
	return strings.ToLower(text) != "true"
}

// TestPullRequest should we test performing a pull request on the repo
func (t *TestOptions) TestPullRequest() bool {
	text := os.Getenv("JX_DISABLE_TEST_PULL_REQUEST")
	return strings.ToLower(text) != "true"
}

// ShouldTestPipelineActivityUpdate should we make sure the build controller is updating the PipelineActivity
func (t *TestOptions) ShouldTestPipelineActivityUpdate() bool {
	return strings.ToLower(DisablePipelineActivityCheck) != "true"
}

// WaitForFirstRelease should we wait for first release to complete before trying a pull request
func (t *TestOptions) WaitForFirstRelease() bool {
	text := os.Getenv("JX_DISABLE_WAIT_FOR_FIRST_RELEASE")
	return strings.ToLower(text) != "true"
}

// WeShouldTestChatOpsCommands should we test prow ChatOps commands
func (t *TestOptions) WeShouldTestChatOpsCommands() bool {
	return strings.ToLower(EnableChatOpsTests) == "true"
}

// ExpectUrlReturns expects that the given URL returns the given status code within the given time period
func (t *TestOptions) ExpectUrlReturns(url string, expectedStatusCode int, maxDuration time.Duration) error {
	lastLoggedStatus := -1
	f := func() error {
		skipVerify := false
		if strings.ToLower(InsecureURLSkipVerify) == "true" {
			skipVerify = true
		}
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipVerify,
			},
		}
		var httpClient = &http.Client{
			Timeout:   time.Second * 30,
			Transport: transport,
		}
		response, err := httpClient.Get(url)
		if err != nil {
			return err
		}
		actualStatusCode := response.StatusCode
		if actualStatusCode != lastLoggedStatus {
			lastLoggedStatus = actualStatusCode
			utils.LogInfof("Invoked %s and got return code: %s\n", util.ColorInfo(url), util.ColorInfo(strconv.Itoa(actualStatusCode)))
		}
		if actualStatusCode == expectedStatusCode {
			return nil
		}
		return fmt.Errorf("invalid HTTP status code for %s expected %d but got %d", url, expectedStatusCode, actualStatusCode)
	}
	return RetryExponentialBackoff(maxDuration, f)
}

func (t *TestOptions) CreateChatOpsCommands(commands []string) error {
	gitProvider, err := t.GetGitProvider()
	if err != nil {
		return err
	}

	utils.LogInfof("successfully create git provider of kind %s", gitProvider.Kind())

	return nil
}

// CreateIssueAndAssignToUser creates an issue on the configure git provider and assigns it to a user.
func (t *TestOptions) CreateIssueAndAssignToUserWithChatOpsCommand(issue *gits.GitIssue, provider gits.GitProvider) error {

	createdIssue, err := provider.CreateIssue(issue.Owner, issue.Repo, issue)
	if err != nil {
		return err
	}

	utils.LogInfof("created issue with number %d\n", *createdIssue.Number)

	err = provider.CreateIssueComment(
		issue.Owner,
		issue.Repo,
		*createdIssue.Number,
		fmt.Sprintf("/assign %s", provider.CurrentUsername()),
	)
	if err != nil {
		return err
	}
	utils.LogInfof("create issue comment on issue %d\n", *createdIssue.Number)

	createdIssue.Owner = issue.Owner
	createdIssue.Repo = issue.Repo

	return t.ExpectThatIssueIsAssignedToUser(provider, createdIssue, provider.CurrentUsername())

}

// ExpectThatIssueIsAssignedToUser returns an error if
func (t *TestOptions) ExpectThatIssueIsAssignedToUser(provider gits.GitProvider, issue *gits.GitIssue, username string) error {
	f := func() error {
		fetchedIssue, err := provider.GetIssue(issue.Owner, issue.Repo, *issue.Number)
		if err != nil {
			return err
		}

		if fetchedIssue == nil {
			return fmt.Errorf("fetched issue is nil but did not throw an error")
		}

		for _, assignee := range fetchedIssue.Assignees {
			if assignee.Login == username {
				return nil
			}
		}

		return fmt.Errorf("user was not found in issue assignees")
	}
	return RetryExponentialBackoff(TimeoutProwActionWait, f)
}

// MostRecentOpenPullRequestForOwnerAndRepo returns the most recently opened pull request for a given owner/repo. If
// there aren't any open PRs, it will return nil.
func (t *TestOptions) MostRecentOpenPullRequestForOwnerAndRepo(provider gits.GitProvider, owner string, repo string) (*gits.GitPullRequest, error) {
	pullRequests, err := provider.ListOpenPullRequests(owner, repo)
	if err != nil {
		return nil, err
	}
	if len(pullRequests) < 1 {
		return nil, fmt.Errorf("no open pull requests found for %s/%s", owner, repo)
	}
	sort.SliceStable(pullRequests, func(i, j int) bool {
		iNum := 0
		jNum := 0
		if pullRequests[i].Number != nil {
			iNum = *pullRequests[i].Number
		}
		if pullRequests[j].Number != nil {
			jNum = *pullRequests[j].Number
		}
		return iNum > jNum
	})

	// The first element in the slice is the open PR with the highest number.
	return pullRequests[0], nil
}

// ApprovePullRequest attempts to /approve a PR with the given approver git provider, then verify the label is there with the default provider
func (t *TestOptions) ApprovePullRequest(defaultProvider gits.GitProvider, approverProvider gits.GitProvider, pullRequest *gits.GitPullRequest) error {
	err := approverProvider.AddPRComment(pullRequest, "/approve")
	if err != nil {
		return err
	}

	return t.ExpectThatPullRequestHasLabel(defaultProvider, *pullRequest.Number, pullRequest.Owner, pullRequest.Repo, "approved")
}

// AttemptToLGTMOwnPullRequest return an error if the /lgtm fails to add the lgtm label to PR
func (t *TestOptions) AttemptToLGTMOwnPullRequest(provider gits.GitProvider, owner string, repo string) error {
	pullRequest, err := t.MostRecentOpenPullRequestForOwnerAndRepo(provider, owner, repo)
	if err != nil {
		return err
	}

	err = provider.AddPRComment(pullRequest, "/lgtm")
	if err != nil {
		return err
	}

	repoStruct := &gits.GitRepository{
		Name:         repo,
		Organisation: owner,
	}
	updatedPullRequest, err := provider.GetPullRequest(owner, repoStruct, *pullRequest.Number)
	if err != nil {
		return err
	}

	return t.ExpectThatPullRequestHasCommentWithText(provider, updatedPullRequest, "you cannot LGTM your own PR.")
}

// ExpectThatPullRequestHasCommentWithText returns an error if the PR does not have a comment with the specified text
func (t *TestOptions) ExpectThatPullRequestHasCommentWithText(provider gits.GitProvider, pullRequest *gits.GitPullRequest, commentText string) error {
	f := func() error {
		userAuth := provider.UserAuth()

		scmClient, err := scmFactory.NewClient(provider.Kind(), provider.ServerURL(), userAuth.ApiToken)
		if err != nil {
			return err
		}

		repoString := fmt.Sprintf("%s/%s", pullRequest.Owner, pullRequest.Repo)
		comments, _, err := scmClient.PullRequests.ListComments(context.Background(), repoString, *pullRequest.Number, scm.ListOptions{})

		for _, comment := range comments {
			if strings.Contains(comment.Body, commentText) {
				return nil
			}
		}
		return fmt.Errorf("comment text not found in PR")
	}

	return RetryExponentialBackoff(TimeoutProwActionWait, f)
}

// AddHoldLabelToPullRequestWithChatOpsCommand returns an error of the command fails to add the do-not-merge/hold label
func (t *TestOptions) AddHoldLabelToPullRequestWithChatOpsCommand(provider gits.GitProvider, owner, repo string) error {
	pullRequest, err := t.MostRecentOpenPullRequestForOwnerAndRepo(provider, owner, repo)
	if err != nil {
		return err
	}

	err = provider.AddPRComment(pullRequest, "/hold")
	if err != nil {
		return err
	}

	return t.ExpectThatPullRequestHasLabel(provider, *pullRequest.Number, owner, repo, "do-not-merge/hold")
}

// AddWIPLabelToPullRequestByUpdatingTitle adds the WIP label by adding WIP to a pull request's title
func (t *TestOptions) AddWIPLabelToPullRequestByUpdatingTitle(provider gits.GitProvider, owner, repo string) error {
	pullRequest, err := t.MostRecentOpenPullRequestForOwnerAndRepo(provider, owner, repo)
	if err != nil {
		return err
	}

	pullRequestArgs := &gits.GitPullRequestArguments{
		Title: fmt.Sprintf("WIP %s", pullRequest.Title),
		GitRepository: &gits.GitRepository{
			Organisation: pullRequest.Owner,
			Name:         pullRequest.Repo,
		},
	}
	updatedPullRequest, err := provider.UpdatePullRequest(pullRequestArgs, *pullRequest.Number)
	if err != nil {
		return err
	}

	return t.ExpectThatPullRequestHasLabel(provider, *updatedPullRequest.Number, owner, repo, "do-not-merge/work-in-progress")
}

// ExpectThatPullRequestHasLabel returns an error if the PR does not have the specified label
func (t *TestOptions) ExpectThatPullRequestHasLabel(provider gits.GitProvider, pullRequestNumber int, owner, repo, label string) error {
	f := func() error {
		repoStruct := &gits.GitRepository{
			Name:         repo,
			Organisation: owner,
		}
		pullRequest, err := provider.GetPullRequest(owner, repoStruct, pullRequestNumber)
		if err != nil {
			return err
		}
		if len(pullRequest.Labels) < 1 {
			return fmt.Errorf("the pull request has no labels")
		}
		for _, l := range pullRequest.Labels {
			if *l.Name == label {
				return nil
			}
		}
		return fmt.Errorf("the pull request does not have the specified label: %s", label)
	}

	return RetryExponentialBackoff(TimeoutProwActionWait, f)
}

func (t *TestOptions) WaitForCreatedPullRequestToMerge(provider gits.GitProvider, prCreateOutput string) {
	createdPR, err := parsers.ParseJxCreatePullRequestFromFullLog(prCreateOutput)
	Expect(err).ShouldNot(HaveOccurred())

	t.WaitForPullRequestToMerge(provider, createdPR.Owner, createdPR.Repository, createdPR.PullRequestNumber, createdPR.Url)
}

func (t *TestOptions) WaitForPullRequestToMerge(provider gits.GitProvider, owner string, repo string, prNumber int, prURL string) {
	repoStruct := &gits.GitRepository{
		Name:         repo,
		Organisation: owner,
	}
	waitForMergeFunc := func() error {
		pr, err := provider.GetPullRequest(owner, repoStruct, prNumber)
		if err != nil {
			utils.LogInfof("WARNING: Error getting pull request: %s\n", err)
			return err
		}
		if pr == nil {
			err = fmt.Errorf("got a nil PR for %s", prURL)
			utils.LogInfof("WARNING: %s\n", err)
			return err
		}
		isMerged := pr.Merged
		if isMerged != nil && *isMerged {
			return nil
		} else {
			err = fmt.Errorf("PR %s not yet merged", prURL)
			utils.LogInfof("WARNING: %s, sleeping and retrying\n", err)
			return err
		}
	}

	err := RetryExponentialBackoff(TimeoutUrlReturns, waitForMergeFunc)
	Expect(err).ShouldNot(HaveOccurred())
}
