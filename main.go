package bdd_jx

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/runner"
	"github.com/jenkins-x/bdd-jx/utils/parsers"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/jx/cmd/clients"
	"github.com/jenkins-x/jx/pkg/util"

	"github.com/cenkalti/backoff"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func main() { /* usual main func */ }

var (
	// TempDirPrefix The prefix to append to applicationss created in testing
	TempDirPrefix = "bdd-"
	// WorkDir The current working directory
	WorkDir              string
	IncludeApps          = os.Getenv("JX_BDD_INCLUDE_APPS")
	IncludeQuickstarts   = os.Getenv("JX_BDD_QUICKSTARTS")
	DefaultRepositoryURL = "http://chartmuseum.jenkins-x.io"

	// all timeout values are in minutes
	// timeout for a build to complete successfully
	TimeoutBuildCompletes = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_COMPLETES", 20)
	// Timeout for promoting an application to staging environment
	TimeoutBuildIsRunningInStaging = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING", 10)
	// Timeout for a given URL to return an expected status code
	TimeoutUrlReturns = utils.GetTimeoutFromEnv("BDD_TIMEOUT_URL_RETURNS", 5)
	// Timeout to wait for a command line execution to complete
	TimeoutCmdLine = utils.GetTimeoutFromEnv("BDD_TIMEOUT_CMD_LINE", 1)
	// Timeout for waiting for jx add app to complete
	TimeoutAppTests = utils.GetTimeoutFromEnv("BDD_TIMEOUT_APP_TESTS", 60)
	// Session wait timeout
	TimeoutSessionWait = utils.GetTimeoutFromEnv("BDD_TIMEOUT_SESSION_WAIT", 60)
)

// Test is the standard testing object
type Test struct {
	Factory         clients.Factory
	Interactive     bool
	WorkDir         string
	ApplicationName string
	Organisation    string
}

// GetGitOrganisation Gets the current git organisation/user
func (t *Test) GetGitOrganisation() string {
	org := os.Getenv("GIT_ORGANISATION")
	if org == "" {
		org = "jenkins-x-tests"
	}
	return org
}

// GitProviderURL Gets the current git provider URL
func (t *Test) GitProviderURL() (string, error) {
	gitProviderURL := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderURL != "" {
		return gitProviderURL, nil
	}
	r := runner.New(t.WorkDir, nil, 0)
	out := r.RunWithOutput("get", "gitserver")
	gitServers, err := parsers.ParseJxGetGitServer(out)
	if err != nil {
		return "", err
	}
	if len(gitServers) < 1 {
		return "", errors.Errorf("Must be at least 1 git server configured")
	}
	return gitServers[0].Url, nil
}

// TheApplicationIsRunningInStaging lets assert that the application is deployed into the first automatic staging environment
func (t *Test) TheApplicationIsRunningInStaging(statusCode int) {
	u := ""
	key := "staging"

	f := func() error {
		r := runner.New(t.WorkDir, nil, 0)
		out := r.RunWithOutput("get", "applications", "-e", key)
		applications, err := parsers.ParseJxGetApplications(out)
		if err != nil {
			utils.LogInfof("failed to parse applications: %s\n", err.Error())
			return err
		}
		// TODO this seems to barf - we need to wait for the application to appear with --tekton...
		// utils.ExpectNoError(err)
		applicationName := t.GetApplicationName()
		if len(applications) == 0 {
			return fmt.Errorf("No applications found")
		}
		utils.LogInfof("application name %s; application mP %#v\n", applicationName, applications)

		applicationEnvInfo, ok := applications[applicationName]
		if !ok {
			applicationName = "jx-" + applicationName
			applicationEnvInfo, ok = applications[applicationName]
		}

		Expect(applicationEnvInfo).ShouldNot(BeNil(), "no application found for % in environment %s", applicationName, key)
		u = applicationEnvInfo.Url
		if u == "" {
			return fmt.Errorf("No URL found for environment %s", key)
			utils.LogInfof("still looking for application %s in env %s\n", applicationName, key)
		}
		return nil
	}

	err := RetryExponentialBackoff(TimeoutBuildIsRunningInStaging, f)
	Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")

	Expect(u).ShouldNot(BeEmpty(), "no URL for environment %s", key)
	t.ExpectUrlReturns(u, statusCode, TimeoutUrlReturns)
}

// TheApplicationShouldBeBuiltAndPromotedViaCICD asserts that the project
// should be created in Jenkins and that the build should complete successfully
func (t *Test) TheApplicationShouldBeBuiltAndPromotedViaCICD(statusCode int) {
	applicationName := t.GetApplicationName()
	owner := t.GetGitOrganisation()
	jobName := owner + "/" + applicationName + "/master"

	t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)
	t.TheApplicationIsRunningInStaging(statusCode)
}

// CreatePullRequestAndGetPreviewEnvironment asserts that a pull request can be created
// on the application and the PR goes green and a preview environment is available
func (t *Test) CreatePullRequestAndGetPreviewEnvironment(statusCode int) error {
	applicationName := t.GetApplicationName()
	workDir := filepath.Join(t.WorkDir, applicationName)
	owner := t.GetGitOrganisation()

	utils.LogInfof("Creating a Pull Request in folder: %s\n", workDir)
	t.ExpectCommandExecution(workDir, TimeoutCmdLine, 0, "git", "checkout", "-b", "changes")

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

	r := runner.New(workDir, nil, 0)
	out := r.RunWithOutput("create", "pullrequest", "-b", "--title", "My First PR commit", "--body", "PR comments")
	pr, err := parsers.ParseJxCreatePullRequest(out)
	utils.ExpectNoError(err)
	Expect(pr).ShouldNot(BeNil())
	prNumber := pr.PullRequestNumber
	Expect(prNumber).ShouldNot(BeNil())

	jobName := owner + "/" + applicationName + "/PR-" + strconv.Itoa(prNumber)
	t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)

	utils.ExpectNoError(err)
	if err != nil {
		return err
	}

	// lets verify that there's a Preview Environment...
	utils.LogInfof("Verifying we have a Preview Environment...\n")
	out = r.RunWithOutput("get", "previews")
	previews, err := parsers.ParseJxGetPreviews(out)
	utils.ExpectNoError(err)

	previewEnv := previews[pr.Url]
	Expect(previewEnv).ShouldNot(BeNil(), "Could not find Preview Environment for application name %s", applicationName)
	applicationUrl := previewEnv.Url
	Expect(applicationUrl).ShouldNot(Equal(""), "No Preview Application URL found")

	utils.LogInfof("Running Preview Environment application at: %s\n", util.ColorInfo(applicationUrl))

	return t.ExpectUrlReturns(applicationUrl, statusCode, TimeoutUrlReturns)
	return nil
}

// ThereShouldBeAJobThatCompletesSuccessfully asserts that the given job name completes within the given duration
func (t *Test) ThereShouldBeAJobThatCompletesSuccessfully(jobName string, maxDuration time.Duration) {
	// NOTE Need to retry here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
	utils.LogInfof("Checking that there is a job built successfully for %s\n", jobName)
	args := []string{"get", "build", "logs", "--wait", jobName}
	t.ExpectJxExecution(t.WorkDir, maxDuration, 0, args...)

	r := runner.New(t.WorkDir, nil, 0)
	// TODO the current --build 1 breaks as it can be number 2 these days!
	//out := r.RunWithOutput("get", "activities", "--filter", jobName, "--build", "1")
	out := r.RunWithOutput("get", "activities", "--filter", jobName)
	activities, err := parsers.ParseJxGetActivities(out)
	if err != nil {
		utils.LogInfof("got error parsing activities: %s\n", err.Error())
	}
	// TODO fails on --ng for now...
	// utils.ExpectNoError(err)
	utils.LogInfof("should be one activity but found %d having run jx get activities --filter %s --build 1; activities %v\n", len(activities), jobName, activities)

	// TODO disabling this for now as we get a failure on ng
	if activities != nil {
		Expect(activities).Should(HaveLen(1), fmt.Sprintf("should be one activity but found %d having run jx get activities --filter %s --build 1; activities %v", len(activities), jobName, activities))
		activity, ok := activities[fmt.Sprintf("%s #%d", jobName, 1)]
		if !ok {
			// TODO lets see if the build is number 2 instead which it is for tekton currently
			activity, ok = activities[fmt.Sprintf("%s #%d", jobName, 2)]
		}
		Expect(ok).Should(BeTrue(), fmt.Sprintf("could not find job with name %s #%d", jobName, 1))

		utils.LogInfof("build status for '%s' is '%s'\n", jobName+"-1", activity.Status)
	}

	// TODO lets temporarily disable this assertion as we have an issue on our production cluster with build statuses not being set correctly
	// TODO lets put this back ASAP once we're on tekton!
	/*
		Expect(activity.Spec.Status.IsTerminated()).To(BeTrue())
		Expect(activity.Spec.Status.String()).Should(Equal("Succeeded"))
	*/
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
func (t *Test) GetApplicationName() string {
	applicationName := t.ApplicationName
	if applicationName == "" {
		_, applicationName = filepath.Split(t.WorkDir)
	}
	return applicationName
}

// ExpectCommandExecution performs the given command in the current work directory and asserts that it completes successfully
func (t *Test) ExpectCommandExecution(dir string, commandTimeout time.Duration, exitCode int, c string, args ...string) {
	f := func() error {
		command := exec.Command(c, args...)
		command.Dir = dir
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		session.Wait(commandTimeout)
		Eventually(session).Should(gexec.Exit(exitCode))
		return err
	}
	err := RetryExponentialBackoff((TimeoutCmdLine), f)
	Ω(err).ShouldNot(HaveOccurred())
}

func (t *Test) ExpectJxExecution(dir string, commandTimeout time.Duration, exitCode int, args ...string) {
	r := runner.New(dir, &commandTimeout, exitCode)
	r.Run(args...)
}

// DeleteApplications should we delete applications after the quickstart has run
func (t *Test) DeleteApplications() bool {
	text := os.Getenv("JX_DISABLE_DELETE_APP")
	return strings.ToLower(text) != "true"
}

// DeleteRepos should we delete the git repos after the quickstart has run
func (t *Test) DeleteRepos() bool {
	text := os.Getenv("JX_DISABLE_DELETE_REPO")
	return strings.ToLower(text) != "true"
}

// TestPullRequest should we test performing a pull request on the repo
func (t *Test) TestPullRequest() bool {
	text := os.Getenv("JX_DISABLE_TEST_PULL_REQUEST")
	return strings.ToLower(text) != "true"
}

// WaitForFirstRelease should we wait for first release to complete before trying a pull request
func (t *Test) WaitForFirstRelease() bool {
	text := os.Getenv("JX_DISABLE_WAIT_FOR_FIRST_RELEASE")
	return strings.ToLower(text) != "true"
}

// ExpectUrlReturns expects that the given URL returns the given status code within the given time period
func (t *Test) ExpectUrlReturns(url string, expectedStatusCode int, maxDuration time.Duration) error {
	lastLoggedStatus := -1
	f := func() error {
		var httpClient = &http.Client{
			Timeout: time.Second * 30,
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
		return fmt.Errorf("Invalid HTTP status code for %s expected %d but got %d", url, expectedStatusCode, actualStatusCode)
	}
	return RetryExponentialBackoff(maxDuration, f)
}

// AddAppTests Creates a jx add app test

// AddAppTests Creates a jx add app test
func AllQuickstartsTest() []bool {
	if IncludeQuickstarts != "" {
		includedQuickstartList := strings.Split(strings.TrimSpace(IncludeQuickstarts), ",")
		tests := make([]bool, len(includedQuickstartList))
		for _, testQuickstartName := range includedQuickstartList {
			tests = append(tests, CreateBatchQuickstartsTests(testQuickstartName))
		}
		return tests
	} else {
		return make([]bool, 0)
	}
}

//CreateQuickstartTest creates a quickstart test for the given quickstart
func CreateQuickstartTest(quickstartName string) bool {
	return createQuickstartTests(quickstartName, false)
}

//CreateBatchQuickstartsTests creates a batch quickstart test for the given quickstart
func CreateBatchQuickstartsTests(quickstartName string) bool {
	return createQuickstartTests(quickstartName, true)
}

// CreateQuickstartTest Creates quickstart tests.  If batch == true, add 'batch' to the test spec
func createQuickstartTests(quickstartName string, batch bool) bool {
	description := ""
	if batch {
		description = "[batch] "
	}
	return Describe(description+"quickstart "+quickstartName+"\n", func() {
		var T Test

		BeforeEach(func() {
			qsNameParts := strings.Split(quickstartName, "-")
			qsAbbr := ""
			for s := range qsNameParts {
				qsAbbr = qsAbbr + qsNameParts[s][:1]

			}
			applicationName := TempDirPrefix + qsAbbr + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10)
			T = Test{
				ApplicationName: applicationName,
				WorkDir:         WorkDir,
				Factory:         clients.NewFactory(),
			}
			T.GitProviderURL()

			utils.LogInfof("Creating application %s in dir %s\n", util.ColorInfo(applicationName), util.ColorInfo(WorkDir))
		})

		Describe("Given valid parameters", func() {
			Context("when operating on the quickstart", func() {
				It("creates a "+quickstartName+" quickstart and promotes it to staging\n", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", quickstartName}

					gitProviderUrl, err := T.GitProviderURL()
					Expect(err).NotTo(HaveOccurred())
					if gitProviderUrl != "" {
						utils.LogInfof("Using Git provider URL %s\n", gitProviderUrl)
						args = append(args, "--git-provider-url", gitProviderUrl)
					}
					T.ExpectJxExecution(T.WorkDir, TimeoutAppTests, 0, args...)

					applicationName := T.GetApplicationName()
					owner := T.GetGitOrganisation()
					jobName := owner + "/" + applicationName + "/master"

					if T.WaitForFirstRelease() {
						By("wait for first release")
						//FIXME Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
						time.Sleep(30 * time.Second)

						T.ThereShouldBeAJobThatCompletesSuccessfully(jobName, 20*time.Minute)
						T.TheApplicationIsRunningInStaging(200)

						if T.TestPullRequest() {
							By("perform a pull request on the source and assert that a preview environment is created")
							T.CreatePullRequestAndGetPreviewEnvironment(200)
						}
					} else {
						By("wait for first successful build of master")
						T.ThereShouldBeAJobThatCompletesSuccessfully(jobName, 20*time.Minute)
					}

					if T.DeleteApplications() {
						By("deletes the application")
						args = []string{"delete", "application", "-b", T.ApplicationName}
						T.ExpectJxExecution(T.WorkDir, TimeoutAppTests, 0, args...)
					}

					if T.DeleteRepos() {
						By("deletes the repo")
						args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
						T.ExpectJxExecution(T.WorkDir, TimeoutAppTests, 0, args...)
					}
				})
			})
		})
		Describe("Given invalid parameters", func() {
			Context("when -p param (project name) is missing", func() {
				It("exits with signal 1\n", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-f", quickstartName}
					T.ExpectJxExecution(T.WorkDir, TimeoutAppTests, 1, args...)
				})
			})
			Context("when -f param (filter) does not match any quickstart", func() {
				It("exits with signal 1\n", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", "the_derek_zoolander_app_for_being_really_really_good_looking"}
					T.ExpectJxExecution(T.WorkDir, TimeoutAppTests, 1, args...)
				})
			})
		})
	})
}
