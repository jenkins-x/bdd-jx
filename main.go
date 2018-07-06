package bdd_jx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

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
	fmt.Fprintf(GinkgoWriter, "Checking that there is a job built successfully for %s\n", jobName)
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

func CreateQuickstartTests(quickstartName string) bool {
	return Describe("quickstart "+quickstartName+"\n", func() {
		var T Test

		BeforeEach(func() {
			T = Test{
				AppName: TempDirPrefix + quickstartName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir: WorkDir,
				Factory: cmd.NewFactory(),
			}

		})

		Describe("Given valid parameters", func() {
			Context("when operating on the quickstart", func() {
				It("creates a "+quickstartName+" quickstart and promotes it to staging\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.AppName, "-f", quickstartName}
					command := exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
					e := T.TheApplicationShouldBeBuiltAndPromotedViaCICD()
					Expect(e).NotTo(HaveOccurred())

					if T.DeleteApps() {
						By("deletes the app")
						fullAppName := T.GetGitOrganisation() + "/" + T.AppName
						args = []string{"delete", "app", "-b", fullAppName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(1 * time.Hour)
						Eventually(session).Should(gexec.Exit(0))
					}

					if T.DeleteRepos() {
						By("deletes the repo")
						args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.AppName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(1 * time.Hour)
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
					command := exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(1))
				})
			})
			Context("when -f param (filter) does not match any quickstart", func() {
				It("exits with signal 1\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.AppName, "-f", "the_derek_zoolander_app_for_being_really_really_good_looking"}
					command := exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})
}
