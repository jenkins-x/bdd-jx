package apps

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"

	"github.com/google/go-github/v28/github"
	"github.com/jenkins-x/bdd-jx/test/helpers"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	"github.com/jenkins-x/jx/pkg/gits"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
		t.SetGitHubToken()
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
		var addAppJobName string
		var deleteAppJobName string
		It("ensure UI is not installed", func() {
			pr, err := t.GetPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Add %s %s", uiAppName, uiAppVersion))
			Expect(err).ShouldNot(HaveOccurred())
			Expect(pr).Should(BeNil())
		})

		It("install UI via 'jx add app'", func() {
			By("installing the app")
			addAppJobName = fmt.Sprintf("%s/%s/master #%s", gitInfo.Organisation, gitInfo.Name, t.NextBuildNumber(gitInfo))
			args := []string{"add", "app", uiAppName, "--version", uiAppVersion, "--repository=https://charts.cloudbees.com/cjxd/cloudbees", "--auto-merge"}
			out := t.ExpectJxExecutionWithOutput(t.WorkDir, timeoutAppTests, 0, args...)

			provider, err := t.GetGitProvider()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(provider).Should(BeNil())

			t.WaitForCreatedPRToMerge(provider, out)

			By("waiting for the build to complete")
			t.TailBuildLog(addAppJobName, helpers.TimeoutBuildCompletes)
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

		It("uninstall UI", func() {
			deleteAppJobName = fmt.Sprintf("%s/%s/master #%s", gitInfo.Organisation, gitInfo.Name, t.NextBuildNumber(gitInfo))
			args := []string{"delete", "app", uiAppName, "--auto-merge"}
			out := t.ExpectJxExecutionWithOutput(t.WorkDir, timeoutAppTests, 0, args...)

			provider, err := t.GetGitProvider()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(provider).Should(BeNil())

			t.WaitForCreatedPRToMerge(provider, out)

			By("waiting for the build to complete")
			t.TailBuildLog(deleteAppJobName, helpers.TimeoutBuildCompletes)
		})
	})
}
