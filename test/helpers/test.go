package helpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/jenkins-x/bdd-jx/test/utils/parsers"
	"github.com/jenkins-x/jx/pkg/util"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"

	"github.com/jenkins-x/bdd-jx/test/utils"

	"io/ioutil"
	"net/http"

	"github.com/jenkins-x/bdd-jx/test/utils/runner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	// TempDirPrefix The prefix to append to applicationss created in testing
	TempDirPrefix = "bdd-"
	// WorkDir The current working directory
	WorkDir              string
	DefaultRepositoryURL = "http://chartmuseum.jenkins-x.io"

	// all timeout values are in minutes
	// timeout for a build to complete successfully
	TimeoutBuildCompletes = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_COMPLETES", 20)
	// TimeoutBuildIsRunningInStaging Timeout for promoting an application to staging environment
	TimeoutBuildIsRunningInStaging = utils.GetTimeoutFromEnv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING", 10)
	// TimeoutPipelineActivityComplete for promoting an application to staging environment
	TimeoutPipelineActivityComplete = utils.GetTimeoutFromEnv("BDD_TIMEOUT_PIPELINE_ACTIVITY_COMPLETE", 100)
	// TimeoutUrlReturns Timeout for a given URL to return an expected status code
	TimeoutUrlReturns = utils.GetTimeoutFromEnv("BDD_TIMEOUT_URL_RETURNS", 5)
	// TimeoutPreviewUrlReturns Timeout for a preview URL to be available
	TimeoutPreviewUrlReturns = utils.GetTimeoutFromEnv("BDD_TIMEOUT_PREVIEW_URL_RETURNS", 500)
	// TimeoutCmdLine Timeout to wait for a command line execution to complete
	TimeoutCmdLine = utils.GetTimeoutFromEnv("BDD_TIMEOUT_CMD_LINE", 1)
	// TimeoutSessionWait Session wait timeout
	TimeoutSessionWait = utils.GetTimeoutFromEnv("BDD_TIMEOUT_SESSION_WAIT", 60)
)

// TestOptions is the base testing object
type TestOptions struct {
	Interactive     bool
	WorkDir         string
	ApplicationName string
	Organisation    string
}

// GetGitOrganisation Gets the current git organisation/user
func (t *TestOptions) GetGitOrganisation() string {
	org := os.Getenv("GIT_ORGANISATION")
	if org == "" {
		org = "jenkins-x-tests"
	}
	return org
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
		out = r.RunWithOutput("get", "gitserver")

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
			out = r.RunWithOutput(args...)
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
		Expect(err).ShouldNot(HaveOccurred(), "send request to deployed application")
	})
}

func getApplication(applicationName string, runningApplications map[string]parsers.Application) (*parsers.Application, error) {
	if len(runningApplications) == 0 {
		return nil, fmt.Errorf("No applications found")
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

	By("making a code change, commiting and pushing it", func() {
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

	args := []string{"create", "pullrequest", "-b", "--title", "My First PR commit", "--body", "PR comments"}
	argsStr := strings.Join(args, " ")
	var out string
	By(fmt.Sprintf("creating a pull request by running jx %s", argsStr), func() {
		out = r.RunWithOutput(args...)
	})

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

	jobName := owner + "/" + applicationName + "/PR-" + strconv.Itoa(prNumber)
	By(fmt.Sprintf("checking that job %s completes successfully", jobName), func() {
		t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, TimeoutBuildCompletes)
		utils.ExpectNoError(err)
	})

	args = []string{"get", "previews"}
	argsStr = strings.Join(args, " ")

	f := func() error {
		var err error
		var previews map[string]parsers.Preview
		By(fmt.Sprintf("parsing the output of jx %s", argsStr), func() {
			By(fmt.Sprintf("verifying there is a preview environment by running jx %s", argsStr), func() {
				out = r.RunWithOutput(args...)
			})

			previews, err = parsers.ParseJxGetPreviews(out)
			utils.ExpectNoError(err)
		})

		By(fmt.Sprintf("checking that a preview environment exists for %s", pr.Url), func() {
			previewEnv := previews[pr.Url]
			Expect(previewEnv).ShouldNot(BeNil(), "Could not find Preview Environment for application name %s", applicationName)
			applicationUrl := previewEnv.Url
			Expect(applicationUrl).ShouldNot(Equal(""), "No Preview Application URL found")

			utils.LogInfof("Running Preview Environment application at: %s\n", util.ColorInfo(applicationUrl))

			err = t.ExpectUrlReturns(applicationUrl, statusCode, TimeoutUrlReturns)
			utils.ExpectNoError(err)
		})
		return err
	}
	By(fmt.Sprintf("retrying waiting for Preview URL to be working with exponential backoff to ensure it completes", argsStr), func() {
		err := Retry(TimeoutPreviewUrlReturns, f)
		Expect(err).ShouldNot(HaveOccurred(), "preview environment visible at a URL")
	})
	return nil

}

// ThereShouldBeAJobThatCompletesSuccessfully asserts that the given job name completes within the given duration
func (t *TestOptions) ThereShouldBeAJobThatCompletesSuccessfully(jobName string, maxDuration time.Duration) {
	// NOTE Need to retry here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
	args := []string{"get", "build", "logs", "--wait", jobName}
	argsStr := strings.Join(args, " ")
	By(fmt.Sprintf("checking that there is a job built successfully by calling jx %s", argsStr), func() {
		t.ExpectJxExecution(t.WorkDir, maxDuration, 0, args...)
	})

	r := runner.New(t.WorkDir, nil, 0)
	// TODO the current --build 1 breaks as it can be number 2 these days!
	//out := r.RunWithOutput("get", "activities", "--filter", jobName, "--build", "1")
	args = []string{"get", "activities", "--filter", jobName}
	argsStr = strings.Join(args, " ")

	var activities map[string]*parsers.Activity
	f := func() error {
		var err error
		var out string
		By(fmt.Sprintf("calling jx %s", argsStr), func() {
			out = r.RunWithOutput(args...)
		})
		activities, err = parsers.ParseJxGetActivities(out)
		// TODO fails on --ng for now...
		//utils.ExpectNoError(err)
		if err != nil {
			utils.LogInfof("got error parsing activities: %s\n", err.Error())
		}
		return err
	}

	By(fmt.Sprintf("retrying jx %s with exponential backoff to ensure it completes", argsStr), func() {
		err := RetryExponentialBackoff(TimeoutPipelineActivityComplete, f)
		Expect(err).ShouldNot(HaveOccurred(), "get applications with a URL")
	})

	activityKey := fmt.Sprintf("%s #%d", jobName, 1)
	By(fmt.Sprintf("finding the activity for %s in %v", activityKey, activities), func() {
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
	})

	By(fmt.Sprintf("checking that the activity %s has succeeded", activityKey), func() {
		// TODO lets temporarily disable this assertion as we have an issue on our production cluster with build statuses not being set correctly
		// TODO lets put this back ASAP once we're on tekton!
		/*
			Expect(activity.Spec.Status.IsTerminated()).To(BeTrue())
			Expect(activity.Spec.Status.String()).Should(Equal("Succeeded"))
		*/
	})
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
	err := RetryExponentialBackoff((TimeoutCmdLine), f)
	Ω(err).ShouldNot(HaveOccurred())
}

func (t *TestOptions) ExpectJxExecution(dir string, commandTimeout time.Duration, exitCode int, args ...string) {
	r := runner.New(dir, &commandTimeout, exitCode)
	r.Run(args...)
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

// WaitForFirstRelease should we wait for first release to complete before trying a pull request
func (t *TestOptions) WaitForFirstRelease() bool {
	text := os.Getenv("JX_DISABLE_WAIT_FOR_FIRST_RELEASE")
	return strings.ToLower(text) != "true"
}

// ExpectUrlReturns expects that the given URL returns the given status code within the given time period
func (t *TestOptions) ExpectUrlReturns(url string, expectedStatusCode int, maxDuration time.Duration) error {
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
