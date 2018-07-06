package bdd_jx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	"github.com/onsi/ginkgo"

	"github.com/jenkins-x/bdd-jx/jenkins"
	"github.com/jenkins-x/golang-jenkins"
)

func main() { /* usual main func */ }

var (
	TempDirPrefix = "bdd-test-"
	WorkDir       string
)

type Test struct {
	Factory       cmd.Factory
	JenkinsClient *gojenkins.Jenkins
	Interactive   bool
	WorkDir       string
	AppName       string
	Organisation  string
}

func (t *Test) GetGitOrganisation() string {
	org := os.Getenv("GIT_ORGANISATION")
	if org == "" {
		org = "jenkins-x-tests"
	}
	return org
}

func (t *Test) GitProviderURL() (string, error) {
	gitProviderURL := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderURL != "" {
		return gitProviderURL, nil
	}
	// find the default load the default one from the current ~/.jx/jenkinsAuth.yaml
	authConfigSvc, err := t.Factory.CreateAuthConfigService("~/.jx/jenkinsAuth.yaml")
	if err != nil {
		return "", err
	}
	config := authConfigSvc.Config()
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

// TheApplicationShouldBeBuiltAndPromotedViaCICD asserts that the project
// should be created in Jenkins and that the build should complete successfully
func (t *Test) TheApplicationShouldBeBuiltAndPromotedViaCICD() error {
	appName := t.AppName
	if appName == "" {
		_, appName = filepath.Split(t.WorkDir)
	}
	owner := t.GetGitOrganisation()
	jobName := owner + "/" + appName + "/master"

	o := cmd.CommonOptions{
		Factory: t.Factory,
	}
	if t.JenkinsClient == nil {
		client, err := o.JenkinsClient()
		if err != nil {
			return err
		}
		t.JenkinsClient = client
	}
	fmt.Fprintf(ginkgo.GinkgoWriter, "Checking that there is a job built successfully for %s\n", jobName)
	return jenkins.ThereShouldBeAJobThatCompletesSuccessfully(jobName, t.JenkinsClient)
}

// DeleteApps should we delete apps after the quickstart has run
func (t *Test) DeleteApps() bool {
	text := os.Getenv("JX_DISABLE_DELETE_APP")
	return strings.ToLower(text) != "true"
}

// DeleteApps should we delete the git repos after the quickstart has run
func (t *Test) DeleteRepos() bool {
	text := os.Getenv("JX_DISABLE_DELETE_REPO")
	return strings.ToLower(text) != "true"
}
