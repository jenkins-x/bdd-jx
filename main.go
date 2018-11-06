package bdd_jx

import (
	"fmt"
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
	v1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() { /* usual main func */ }

var (
	// TempDirPrefix The prefix to append to apps created in testing
	TempDirPrefix = "bdd-"
	// WorkDir The current working directory
	WorkDir string
)

// Test is the standard testing object
type Test struct {
	Factory      cmd.Factory
	Interactive  bool
	WorkDir      string
	AppName      string
	Organisation string
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
	authConfigSvc, err := t.Factory.CreateAuthConfigService("~/.jx/gitAuth.yaml")
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
func (t *Test) TheApplicationShouldBeBuiltAndPromotedViaCICD() {
	appName := t.GetAppName()
	owner := t.GetGitOrganisation()
	jobName := owner + "/" + appName + "/master"

	t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, 10*time.Minute)
}

// CreatePullRequestAndGetPreviewEnvironment asserts that a pull request can be created
// on the application and the PR goes green and a preview environment is available
func (t *Test) CreatePullRequestAndGetPreviewEnvironment(statusCode int) error {
	appName := t.GetAppName()
	workDir := filepath.Join(t.WorkDir, appName)
	owner := t.GetGitOrganisation()

	fmt.Printf("Creating a Pull Request in folder: %s\n", workDir)

	t.ExpectCommandExecution(workDir, time.Minute, 0, "git", "checkout", "-b", "changes")

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

	jobName := owner + "/" + appName + "/PR-" + strconv.Itoa(*prNumber)

	t.ThereShouldBeAJobThatCompletesSuccessfully(jobName, 10*time.Minute)

	Expect(err).ShouldNot(HaveOccurred())
	if err != nil {
		return err
	}

	// lets verify that there's a Preview Environment...
	fmt.Printf("Verifying we have a Preview Environment...\n")
	jxClient, ns, err := o.JXClientAndDevNamespace()
	Expect(err).ShouldNot(HaveOccurred())

	envList, err := jxClient.JenkinsV1().Environments(ns).List(metav1.ListOptions{})
	Expect(err).ShouldNot(HaveOccurred())

	var previewEnv *v1.Environment
	for _, env := range envList.Items {
		spec := &env.Spec
		if spec.Kind == v1.EnvironmentKindTypePreview {
			if spec.PreviewGitSpec.ApplicationName == appName {
				copy := env
				previewEnv = &copy
			}
		}
	}
	Expect(previewEnv).ShouldNot(BeNil(), "Could not find Preview Environment in namespace %s for app name %s", ns, appName)
	if previewEnv != nil {
		appUrl := previewEnv.Spec.PreviewGitSpec.ApplicationURL
		Expect(appUrl).ShouldNot(Equal(""), "No Preview Application URL found")

		fmt.Printf("Running Preview Environment application at: %s\n", util.ColorInfo(appUrl))

		return t.ExpectUrlReturns(appUrl, statusCode, time.Minute*5)
	} else {
		fmt.Printf("No Preview Environment found in namespace %s for application: %s\n", ns, appName)
	}
	return nil
}

// ThereShouldBeAJobThatCompletesSuccessfully asserts that the given job name completes within the given duration
func (t *Test) ThereShouldBeAJobThatCompletesSuccessfully(jobName string, maxDuration time.Duration) {
	// NOTE Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
	time.Sleep(20 * time.Second)
	fmt.Fprintf(GinkgoWriter, "Checking that there is a job built successfully for %s\n", jobName)
	t.ExpectCommandExecution(t.WorkDir, (time.Minute * 10), 0, "jx", "get", "build", "logs", jobName)
}

// RetryExponentialBackoff retries the given function up to the maximum duration
func (t *Test) RetryExponentialBackoff(maxDuration time.Duration, f func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = maxDuration
	exponentialBackOff.Reset()
	err := backoff.Retry(f, exponentialBackOff)
	return err
}

// GetAppName gets the app name for the current test case
func (t *Test) GetAppName() string {
	appName := t.AppName
	if appName == "" {
		_, appName = filepath.Split(t.WorkDir)
	}
	return appName
}

// ExpectCommandExecution performs the given command in the current work directory and asserts that it completes successfully
func (t *Test) ExpectCommandExecution(dir string, commandTimeout time.Duration, exitCode int, c string, args ...string) {
	command := exec.Command(c, args...)
	command.Dir = dir
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	session.Wait(commandTimeout)
	Eventually(session).Should(gexec.Exit(exitCode))
}

// DeleteApps should we delete apps after the quickstart has run
func (t *Test) DeleteApps() bool {
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
			fmt.Printf("Invoked %s and got return code: %s\n", util.ColorInfo(url), util.ColorInfo(strconv.Itoa(actualStatusCode)))
		}
		if actualStatusCode == expectedStatusCode {
			return nil
		}
		return fmt.Errorf("Invalid HTTP status code for %s expected %d but got %d", url, expectedStatusCode, actualStatusCode)
	}
	return t.RetryExponentialBackoff(maxDuration, f)
}

// CreateQuickstartTests Creates quickstart tests
func CreateQuickstartTests(quickstartName string) bool {
	return Describe("quickstart "+quickstartName+"\n", func() {
		var T Test

		BeforeEach(func() {
			T = Test{
				AppName: TempDirPrefix + quickstartName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir: WorkDir,
				Factory: cmd.NewFactory(),
			}
			T.GitProviderURL()
		})

		commandTimeout := 1 * time.Hour
		Describe("Given valid parameters", func() {
			Context("when operating on the quickstart", func() {
				It("creates a "+quickstartName+" quickstart and promotes it to staging\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.AppName, "-f", quickstartName}

					gitProviderUrl, err := T.GitProviderURL()
					Expect(err).NotTo(HaveOccurred())
					if gitProviderUrl != "" {
						fmt.Fprintf(GinkgoWriter, "Using Git provider URL %s\n", gitProviderUrl)
						args = append(args, "--git-provider-url", gitProviderUrl)
					}
					command := exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(commandTimeout)
					Eventually(session).Should(gexec.Exit(0))

					if T.WaitForFirstRelease() {
						By("wait for first release")
						// NOTE Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
						time.Sleep(20 * time.Second)
						T.TheApplicationShouldBeBuiltAndPromotedViaCICD()
					}

					if T.TestPullRequest() {
						By("perform a pull request on the source and assert that a preview environment is created")
						T.CreatePullRequestAndGetPreviewEnvironment(200)
					}

					if T.DeleteApps() {
						By("deletes the app")
						fullAppName := T.GetGitOrganisation() + "/" + T.AppName
						args = []string{"delete", "app", "-b", fullAppName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(commandTimeout)
						Eventually(session).Should(gexec.Exit(0))
					}

					if T.DeleteRepos() {
						By("deletes the repo")
						args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.AppName}
						command = exec.Command(c, args...)
						command.Dir = T.WorkDir
						session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
						Ω(err).ShouldNot(HaveOccurred())
						session.Wait(commandTimeout)
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
					T.ExpectCommandExecution(T.WorkDir, commandTimeout, 1, c, args...)
				})
			})
			Context("when -f param (filter) does not match any quickstart", func() {
				It("exits with signal 1\n", func() {
					c := "jx"
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.AppName, "-f", "the_derek_zoolander_app_for_being_really_really_good_looking"}
					T.ExpectCommandExecution(T.WorkDir, commandTimeout, 1, c, args...)
				})
			})
		})
	})
}
