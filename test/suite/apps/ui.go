package apps

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/jenkins-x/bdd-jx/test/helpers"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	"github.com/jenkins-x/jx/pkg/gits"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

var _ = Describe("Jenkins X UI tests", func() {
	var appTestOptions AppTestOptions

	BeforeEach(func() {
		appTestOptions = AppTestOptions{
			helpers.TestOptions{
				ApplicationName: helpers.TempDirPrefix + "ui" + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         helpers.WorkDir,
			},
		}
		if !appTestOptions.GitOpsEnabled() {
			Skip("Skipping apps tests for UI since they require a gitops setup")
		}
	})

	_ = appTestOptions.UITest()
})

func (t *AppTestOptions) UITest() bool {
	var (
		jxHome       string
		ctx          context.Context
		gitHubClient *github.Client
		gitInfo      *gits.GitRepository
		err          error
	)

	BeforeEach(func() {
		By("setting a temporary JX_HOME directory")
		jxHome, err = ioutil.TempDir("", helpers.TempDirPrefix+"ui-jx-home-")
		Expect(err).ShouldNot(HaveOccurred())

		_ = os.Setenv("JX_HOME", jxHome)
		utils.LogInfo(fmt.Sprintf("Using '%s' as JX_HOME", jxHome))
	})

	BeforeEach(func() {
		By("setting the GitHub token")
		setGitHubToken()
	})

	BeforeEach(func() {
		By("setting up a GitHub client")
		ctx = context.Background()
		gitHubClient = t.GitHubClient()
	})

	BeforeEach(func() {
		By("parsing the gitops dev repo information")
		gitInfo, err = gits.ParseGitURL(t.GitOpsDevRepo())
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(jxHome)
	})

	return Context("UI", func() {
		var uiURL = ""
		It("ensure UI is not installed", func() {
			pr, err := getPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Add %s %s", uiAppName, uiAppVersion))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pr).Should(BeNil())
		})

		It("install UI via 'jx add app'", func() {
			By("installing the app")
			args := []string{"add", "app", uiAppName, "--version", uiAppVersion, "--repository=https://charts.cloudbees.com/cjxd/cloudbees"}
			t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)

			pr, err := getPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Add %s %s", uiAppName, uiAppVersion))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pr).ShouldNot(BeNil())
			Expect(*pr.State).Should(Equal("open"))

			By("merging the install app PR")
			results, _, err := gitHubClient.PullRequests.Merge(ctx, gitInfo.Organisation, gitInfo.Name, *pr.Number, "PR merge", nil)
			Expect(pr).ShouldNot(BeNil())
			Expect(*results.Merged).Should(BeTrue())

			By("waiting for the build to complete")
			jobName := fmt.Sprintf("%s/%s/master #%s", gitInfo.Organisation, gitInfo.Name, t.NextBuildNumber(gitInfo))
			t.TailBuildLog(jobName, helpers.TimeoutBuildCompletes)
		})

		It("ensure UI is installed", func() {
			args := []string{"get", "app", uiAppName}
			t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)
		})

		It("Accessing the UI", func() {
			By("opening port forward")
			port, err := t.GetFreePort()
			Expect(err).ShouldNot(HaveOccurred())

			// even though the app is installed, the deployment might not be ready yet. Let's wait for it.
			t.WaitForDeploymentRollout("jenkins-x-jxui")

			go func() {
				defer GinkgoRecover()

				args := []string{"ui", "-p", strconv.Itoa(port), "-u"}
				command := exec.Command(runner.JxBin(), args...)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Expect(err).ShouldNot(HaveOccurred())
				session.Wait(timeoutAppTests)
				Eventually(session).Should(gexec.Exit())
			}()

			testUI := func() error {
				uiURL = fmt.Sprintf("http://127.0.0.1:%d", port)
				By(fmt.Sprintf("accessing ui on %s", uiURL))
				resp, err := http.Get(uiURL)
				if err != nil {
					return err
				}
				defer func() {
					_ = resp.Body.Close()
				}()
				contents, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				Expect(string(contents)).Should(ContainSubstring("<title>CJXD UI</title>"))
				return nil
			}

			err = helpers.RetryExponentialBackoff(helpers.TimeoutUrlReturns, testUI)
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("Runs smoke tests", func() {
			By("Running smoke tests", func() {
				dir, err := os.Getwd()
				if err != nil {
					fmt.Println(err)
				}

				nodePath, err := exec.LookPath("node")
				if err != nil {
					fmt.Println("Can't find node in your PATH")
				}
				Expect(err).ShouldNot(HaveOccurred())

				command := exec.Command("/bin/sh", "run.sh")
				command.Dir = fmt.Sprintf("%s/ui-smoke", dir)
				command.Env = append(command.Env, fmt.Sprintf("CYPRESS_BASE_URL=%s", uiURL))
				command.Env = append(command.Env, fmt.Sprintf("REPORTS_DIR=%s", os.Getenv("REPORTS_DIR")))
				command.Env = append(command.Env, fmt.Sprintf("PATH=%s:%s", nodePath, os.Getenv("PATH")))
				command.Stderr = GinkgoWriter
				command.Stdout = GinkgoWriter
				err = command.Run()
				if err != nil {
					fmt.Println("Smoke tests failed")
				}
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		It("uninstall UI", func() {
			args := []string{"delete", "app", uiAppName}
			t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)

			pr, err := getPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Delete %s", uiAppName))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pr).ShouldNot(BeNil())

			By("merging the uninstall app PR")
			results, _, err := gitHubClient.PullRequests.Merge(ctx, gitInfo.Organisation, gitInfo.Name, *pr.Number, "PR merge", nil)
			Expect(pr).ShouldNot(BeNil())
			Expect(*results.Merged).Should(BeTrue())

			By("waiting for the build to complete")
			jobName := fmt.Sprintf("%s/%s/master #%s", gitInfo.Organisation, gitInfo.Name, t.NextBuildNumber(gitInfo))
			t.TailBuildLog(jobName, helpers.TimeoutBuildCompletes)
		})
	})
}
