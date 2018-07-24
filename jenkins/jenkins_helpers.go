package jenkins

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/golang-jenkins"
	"github.com/onsi/ginkgo"
)

const (
	maxWaitForBuildToStart     = 40 * time.Second
	maxWaitForBuildToBeCreated = 100 * time.Second
	maxWaitForBuildToComplete  = 400 * time.Minute
)

var jenkinsLogPrefix = utils.Color("\x1b[36m") + "        "

// Is404 returns true if this is a 404 error
func Is404(err error) bool {
	text := fmt.Sprintf("%s", err)
	return strings.HasPrefix(text, "404 ")
}

// TriggerAndWaitForBuildToStart triggers the build and waits for a new Build for the given amount of time
// or returns an error
func TriggerAndWaitForBuildToStart(jenkins *gojenkins.Jenkins, job gojenkins.Job, buildStartWaitTime time.Duration) (result *gojenkins.Build, err error) {
	previousBuildNumber := 0
	previousBuild, err := jenkins.GetLastBuild(job)
	jobUrl := job.Url
	if err != nil {
		if !Is404(err) {
			//return nil, fmt.Errorf("error finding last build for %s due to %v", job.Name, err)
			utils.LogInfof("Warning: error finding previous build for %s due to %v\n", jobUrl, err)
		}
	} else {
		previousBuildNumber = previousBuild.Number
	}
	err = jenkins.Build(job, nil)
	if err != nil {
		if !Is404(err) {
			return nil, fmt.Errorf("error triggering build %s due to %v", jobUrl, err)
		}
	}
	attempts := 0

	// lets wait for a new build to start
	fn := func() (bool, error) {
		buildNumber := 0
		attempts += 1
		build, err := jenkins.GetLastBuild(job)
		if err != nil {
			if !Is404(err) {
				//return nil, fmt.Errorf("error finding last build for %s due to %v", job.Name, err)
				utils.LogInfof("Warning: error finding last build attempt %d for %s due to %v\n", attempts, jobUrl, err)
			}
		} else {
			buildNumber = build.Number
		}
		if previousBuildNumber != buildNumber {
			utils.LogInfof("triggered job %s build #%d\n", jobUrl, buildNumber)
			result = &build
			return true, nil
		}
		return false, nil
	}
	err = gojenkins.Poll(1*time.Second, buildStartWaitTime, fmt.Sprintf("build to start for for %s", jobUrl), fn)
	return
}

// TriggerAndWaitForBuildToStart triggers the build and waits for a new Build then waits for the Build to finish
// or returns an error
func TriggerAndWaitForBuildToFinish(jenkins *gojenkins.Jenkins, job gojenkins.Job, buildStartWaitTime time.Duration, buildFinishWaitTime time.Duration) (*gojenkins.Build, error) {
	build, err := TriggerAndWaitForBuildToStart(jenkins, job, buildStartWaitTime)
	if err != nil {
		return build, err
	}
	if !build.Building {
		return build, nil
	}
	return WaitForBuildToFinish(jenkins, job, build.Number, buildFinishWaitTime)
}

// TriggerAndWaitForBuildToStart triggers the build and waits for a new Build then waits for the Build to finish
// or returns an error
func WaitForBuildToFinish(jenkins *gojenkins.Jenkins, job gojenkins.Job, buildNumber int, buildFinishWaitTime time.Duration) (*gojenkins.Build, error) {
	jobUrl := job.Url
	utils.LogInfof("waiting for job %s build #%d to finish\n", jobUrl, buildNumber)
	time.Sleep(1 * time.Second)
	var result *gojenkins.Build

	fn := func() (bool, error) {
		if result != nil {
			return true, nil
		}
		b, err := jenkins.GetBuild(job, buildNumber)
		if err != nil {
			return false, fmt.Errorf("error finding job %s build #%d status due to %v", jobUrl, buildNumber, err)
		}
		if !b.Building {
			result = &b
			return true, nil
		}
		return false, nil
	}
	writer := utils.NewPrefixWriter(ginkgo.GinkgoWriter, jenkinsLogPrefix)
	logFn := jenkins.TailLogFunc(jenkins.GetBuildURL(job, buildNumber), writer)
	/*
		poller := jenkins.NewLogPoller(jenkins.GetBuildURL(job, buildNumber), os.Stdout)
		logFn := func() (bool, error) {
			return poller.Apply()
		}
	*/
	fns := gojenkins.NewConditionFunc(fn, logFn)
	err := gojenkins.Poll(2*time.Second, buildFinishWaitTime, fmt.Sprintf("job %s build #%d to finish", jobUrl, buildNumber), fns)
	if err == nil && result == nil {
		return result, fmt.Errorf("No build found for job %s", jobUrl)
	}
	return result, err
}

// WaitForBuildLog
func WaitForBuildLog(jenkins *gojenkins.Jenkins, buildURL string, buildFinishWaitTime time.Duration) error {
	utils.LogInfof("waiting for job %s to finish\n", buildURL)
	time.Sleep(1 * time.Second)

	poller := jenkins.NewLogPoller(buildURL, os.Stdout)
	logFn := func() (bool, error) {
		return poller.Apply()
	}
	return gojenkins.Poll(1*time.Second, buildFinishWaitTime, fmt.Sprintf("waiting for job %s to finish\n", buildURL), logFn)
}

// AssertBuildSucceeded asserts that the given build succeeded
func AssertBuildSucceeded(build *gojenkins.Build, jobName string) error {
	if build == nil {
		return fmt.Errorf("No build available for job %s", jobName)
	}
	result := build.Result
	utils.LogInfof("Job %s build %d has result %s\n", jobName, build.Number, result)
	if result == "SUCCESS" {
		return nil
	}
	return fmt.Errorf("Job %s build %d has result %s", jobName, build.Number, result)

}

func GetJobByExpression(jobExpression string, jenkins *gojenkins.Jenkins) (job gojenkins.Job, err error) {
	jobPath := utils.ReplaceEnvVars(jobExpression)

	paths := strings.Split(jobPath, "/")
	fullPath := gojenkins.FullJobPath(paths...)

	job, err = jenkins.GetJobByPath(paths...)
	if err != nil {
		err = fmt.Errorf("Failed to find job %s due to %v", fullPath, err)
	}
	return
}

func ThereShouldBeAJobThatCompletesSuccessfully(jobExpression string, jenkins *gojenkins.Jenkins) error {
	job, err := WaitForJobByExpression(jobExpression, maxWaitForBuildToBeCreated, jenkins)
	if err != nil {
		return err
	}
	lastBuild, err := jenkins.GetLastBuild(job)
	var lastBuildNumber int
	if err != nil {
		if Is404(err) {
			lastBuildNumber = 0
		} else {
			return fmt.Errorf("Failed to find last build for job %s due to %v\n", job.Name, err)
		}
	} else {
		lastBuildNumber = lastBuild.Number
	}
	build, err := WaitForBuildToFinish(jenkins, job, lastBuildNumber, maxWaitForBuildToComplete)
	if err != nil {
		return err
	}
	return AssertBuildSucceeded(build, job.Url)
}

func WaitForJobByExpression(jobExpression string, timeout time.Duration, jenkins *gojenkins.Jenkins) (job gojenkins.Job, err error) {
	jobPath := utils.ReplaceEnvVars(jobExpression)

	paths := strings.Split(jobPath, "/")
	fullPath := gojenkins.FullJobPath(paths...)

	fn := func() (bool, error) {
		job, err = jenkins.GetJobByPath(paths...)
		if err != nil {
			if !Is404(err) {
				err = fmt.Errorf("Failed to find job %s due to %v", fullPath, err)
				return false, err
			}
		} else {
			return true, nil
		}
		return false, nil
	}
	err = gojenkins.Poll(1*time.Second, timeout, fmt.Sprintf("build to be created for %s", fullPath), fn)
	return
}
