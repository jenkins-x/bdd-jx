package bdd_jx

import (
	"flag"
	"fmt"
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/kube"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/jx/pkg/util"

	"github.com/cenkalti/backoff"
	"github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() { /* usual main func */ }

var (
	// TempDirPrefix The prefix to append to applicationss created in testing
	TempDirPrefix = "bdd-"
	// WorkDir The current working directory
	WorkDir              string
	IncludeApps          = flag.String("include-apps", "", "The Jenkins X App names to BDD test")
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
	// Timeout to wait for a jx get build log to return data
	TimeoutGetBuildLog = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_LOG", 10)
	// Timeout for waiting for jx add app to complete
	TimeoutAppTests = utils.GetTimeoutFromEnv("BDD_TIMEOUT_APP_TESTS", 60)
	// Session wait timeout
	TimeoutSessionWait = utils.GetTimeoutFromEnv("BDD_TIMEOUT_SESSION_WAIT", 60)
)

// Test is the standard testing object
type Test struct {
	Factory         cmd.Factory
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
	// find the default load the default one from the current ~/.jx/gitAuth.yaml
	authConfigSvc, err := t.Factory.CreateAuthConfigService("gitAuth.yaml")
	if err != nil {
		return "", err
	}
	config, err := authConfigSvc.LoadConfig()
	if err != nil {
		return "", err
	}
	url := config.CurrentServer
	if url != "" {
		return url, nil
	}
	servers := config.Servers
	if len(servers) == 0 {
		return "", fmt.Errorf("No servers in the ~/.jx/gitAuth.yaml file")
	}
	return servers[0].URL, nil
}

// TheApplicationIsRunningInStaging lets assert that the application is deployed into the first automatic staging environment
func (t *Test) TheApplicationIsRunningInStaging(statusCode int) {
	u := ""
	key := "staging"

	f := func() error {
		o := &cmd.GetApplicationsOptions{
			CommonOptions: cmd.CommonOptions{
				Factory: t.Factory,
				Out:     os.Stdout,
				Err:     os.Stderr,
			},
		}
		err := o.Run()
		Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")
		if err != nil {
			return err
		}

		applicationName := t.GetApplicationName()
		if len(o.Results.Applications) == 0 {
			return fmt.Errorf("No applications found")
		}
		utils.LogInfof("application name %s application map %#v\n", applicationName, o.Results.Applications)

		applicationEnvInfo := o.Results.Applications[applicationName]
		applicationName2 := "jx-" + applicationName
		if applicationEnvInfo == nil {
			applicationEnvInfo = o.Results.Applications[applicationName2]
		}

		if applicationEnvInfo != nil {
			m := applicationEnvInfo[key]
			if m == nil {
				for k := range applicationEnvInfo {
					utils.LogInfof("has environment key %s\n", k)
				}
			}
			Expect(m).ShouldNot(BeNil(), "no ApplicationEnvInfo for key %s", key)
			if m != nil {
				u = m.URL
			}
		}
		if u == "" {
			return fmt.Errorf("No URL found for environment %s", key)
			utils.LogInfo("still looking for application env info url\n")
		}
		return nil
	}
	err := RetryExponentialBackoff(TimeoutBuildIsRunningInStaging, f)
	Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")

	Expect(u).ShouldNot(BeEmpty(), "no ApplicationEnvInfo URL for environment key %s", key)
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

	o := cmd.CreatePullRequestOptions{
		CreateOptions: cmd.CreateOptions{
			CommonOptions: cmd.CommonOptions{
				Factory:   t.Factory,
				Out:       os.Stdout,
				Err:       os.Stderr,
				BatchMode: true,
			},
		},
		Title: "My First PR commit",
		Body:  "PR comments",
		Dir:   workDir,
		Base:  "master",
	}

	err = o.Run()
	pr := o.Results.PullRequest

	Expect(err).ShouldNot(HaveOccurred())
	Expect(pr).ShouldNot(BeNil())
	prNumber := pr.Number
	Expect(prNumber).ShouldNot(BeNil())

	jobName := owner + "/" + applicationName + "/PR-" + strconv.Itoa(*prNumber)

	t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)

	Expect(err).ShouldNot(HaveOccurred())
	if err != nil {
		return err
	}

	// lets verify that there's a Preview Environment...
	utils.LogInfof("Verifying we have a Preview Environment...\n")
	jxClient, ns, err := o.JXClientAndDevNamespace()
	Expect(err).ShouldNot(HaveOccurred())

	envList, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	var previewEnv *v1.Environment
	for _, env := range envList.Items {
		spec := &env.Spec
		if spec.Kind == v1.EnvironmentKindTypePreview {
			if spec.PreviewGitSpec.ApplicationName == applicationName {
				copy := env
				previewEnv = &copy
			}
		}
	}
	Expect(previewEnv).ShouldNot(BeNil(), "Could not find Preview Environment in namespace %s for application name %s", ns, applicationName)
	if previewEnv != nil {
		applicationUrl := previewEnv.Spec.PreviewGitSpec.ApplicationURL
		Expect(applicationUrl).ShouldNot(Equal(""), "No Preview Application URL found")

		utils.LogInfof("Running Preview Environment application at: %s\n", util.ColorInfo(applicationUrl))

		return t.ExpectUrlReturns(applicationUrl, statusCode, TimeoutUrlReturns)
	} else {
		utils.LogInfof("No Preview Environment found in namespace %s for application: %s\n", ns, applicationName)
	}
	return nil
}

// ThereShouldBeAJobThatCompletesSuccessfully asserts that the given job name completes within the given duration
func (t *Test) ThereShouldBeAJobThatCompletesSuccessfully(jobName string, maxDuration time.Duration) {
	// NOTE Need to retry here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
	utils.LogInfof("Checking that there is a job built successfully for %s\n", jobName)
	t.ExpectCommandExecution(t.WorkDir, TimeoutGetBuildLog, 0, "jx", "get", "build", "logs", "--wait", jobName)

	o := cmd.CommonOptions{
			Factory:   t.Factory,
			Out:       os.Stdout,
			Err:       os.Stderr,
			BatchMode: true,
	}

	jxClient, ns, err := o.JXClientAndDevNamespace()
	Expect(err).ShouldNot(HaveOccurred())
	activity, err := jxClient.JenkinsV1().PipelineActivities(ns).Get(kube.ToValidName(jobName + "-1"), metav1.GetOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	utils.LogInfof("build status for '%s' is '%s'", jobName + "-1", activity.Spec.Status.String())

	Expect(activity.Spec.Status.IsTerminated()).To(BeTrue())
	Expect(activity.Spec.Status.String()).Should(Equal("Succeeded"))
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
func AppTests() []bool {
	flag.Parse()
	includedApps := *IncludeApps
	if includedApps != "" {
		includedAppList := strings.Split(strings.TrimSpace(includedApps), ",")
		tests := make([]bool, len(includedAppList))
		for _, testAppName := range includedAppList {
			nameAndVersion := strings.Split(testAppName, ":")
			if len(nameAndVersion) == 2 {
				tests = append(tests, AppTest(nameAndVersion[0], nameAndVersion[1]))
			} else {
				tests = append(tests, AppTest(testAppName, ""))
			}

		}
		return tests
	} else {
		return nil
	}
}

func AppTest(testAppName string, version string) bool {
	return Describe("test app "+testAppName+"\n", func() {
		var T Test

		BeforeEach(func() {
			T = Test{
				ApplicationName: TempDirPrefix + testAppName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         WorkDir,
				Factory:         cmd.NewFactory(),
			}
			T.GitProviderURL()
		})

		_ = T.AddAppTests(testAppName, version)
		_ = T.DeleteAppTests(testAppName)

	})
}

// AddAppTests add app tests
func (t *Test) AddAppTests(testAppName string, version string) bool {
	return Describe("Given valid parameters", func() {
		Context("when running jx add app "+testAppName, func() {
			helmAppName := testAppName + "-" + testAppName
			It("Ensure the app is added\n", func() {
				By("The App resource does not exist before creation\n")
				c := "kubectl"
				args := []string{"get", "app", helmAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 1, c, args...)
				By("Add app exits with signal 0\n")
				c = "jx"
				args = []string{"add", "app", testAppName, "--repository", DefaultRepositoryURL}
				if version != "" {
					args = append(args, "--version", version)
				}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("The App resource exists after creation\n")
				c = "kubectl"
				args = []string{"get", "app", helmAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
			})
		})
	})
}

// DeleteAppTests delete app tests
func (t *Test) DeleteAppTests(testAppName string) bool {
	return Describe("Given valid parameters", func() {
		Context("when running jx delete app "+testAppName, func() {
			helmAppName := testAppName + "-" + testAppName
			It("Ensure it is deleted\n", func() {
				By("The App resource exists before deletion\n")
				c := "kubectl"
				args := []string{"get", "app", helmAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("Delete app exits with signal 0\n")
				c = "jx"
				args = []string{"delete", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("The App resource was removed\n")
				c = "kubectl"
				args = []string{"get", "app", helmAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 1, c, args...)
			})
		})
	})
}

// CreateQuickstartTests Creates quickstart tests
func CreateQuickstartTests(quickstartName string) bool {
	return Describe("quickstart "+quickstartName+"\n", func() {
		var T Test

		BeforeEach(func() {
			applicationName := TempDirPrefix + quickstartName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10)
			T = Test{
				ApplicationName: applicationName,
				WorkDir:         WorkDir,
				Factory:         cmd.NewFactory(),
			}
			T.GitProviderURL()

			utils.LogInfof("Creating application %s in dir %s\n", util.ColorInfo(applicationName), util.ColorInfo(WorkDir))
		})

		Describe("Given valid parameters", func() {
			Context("when operating on the quickstart", func() {
				It("creates a "+quickstartName+" quickstart and promotes it to staging\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", quickstartName}

					gitProviderUrl, err := T.GitProviderURL()
					Expect(err).NotTo(HaveOccurred())
					if gitProviderUrl != "" {
						utils.LogInfof("Using Git provider URL %s\n", gitProviderUrl)
						args = append(args, "--git-provider-url", gitProviderUrl)
					}
					command := exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(TimeoutAppTests)
					Eventually(session).Should(gexec.Exit(0))

					if T.WaitForFirstRelease() {
						By("wait for first release")
						//FIXME Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
						time.Sleep(30 * time.Second)
						T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)
					}

					if T.TestPullRequest() {
						By("perform a pull request on the source and assert that a preview environment is created")
						T.CreatePullRequestAndGetPreviewEnvironment(200)
					}

					if T.DeleteApplications() {
						By("deletes the application")
						args = []string{"delete", "application", "-b", T.ApplicationName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(TimeoutAppTests)
						Eventually(session).Should(gexec.Exit(0))
					}

					if T.DeleteRepos() {
						By("deletes the repo")
						args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(TimeoutAppTests)
						Eventually(session).Should(gexec.Exit(0))
					}
				})
			})
		})
		Describe("Given invalid parameters", func() {
			Context("when -p param (project name) is missing", func() {
				It("exits with signal 1\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-f", quickstartName}
					T.ExpectCommandExecution(T.WorkDir, TimeoutAppTests, 1, c, args...)
				})
			})
			Context("when -f param (filter) does not match any quickstart", func() {
				It("exits with signal 1\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", "the_derek_zoolander_app_for_being_really_really_good_looking"}
					T.ExpectCommandExecution(T.WorkDir, TimeoutAppTests, 1, c, args...)
				})
			})
		})
	})
}
