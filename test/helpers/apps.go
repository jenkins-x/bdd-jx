package helpers

import (
	"github.com/google/go-github/v28/github"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/net/context"
	"os"
	"os/exec"
)

func SetGitHubToken() {
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

func GetPullRequestWithTitle(client *github.Client, ctx context.Context, repoOwner string, repoName string, title string) (*github.PullRequest, error) {
	pullRequestList, _, err := client.PullRequests.List(ctx, repoOwner, repoName, nil)
	if err != nil {
		return nil, err

	}

	var matchingPR *github.PullRequest
	for _, pullRequest := range pullRequestList {
		if *pullRequest.Title == title {
			matchingPR = pullRequest
		}
	}

	return matchingPR, nil
}
