package jxui

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
	"strings"
	"time"
)

type AppTestOptions struct {
	helpers.TestOptions
}

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
		uiURL        string
		err          error
	)

	BeforeEach(func() {
		if jxUiUrl := runner.JxUiUrl(); jxUiUrl != "" {
			uiURL = jxUiUrl
		} else {
			By("setting a temporary JX_HOME directory")
			jxHome, err = ioutil.TempDir("", helpers.TempDirPrefix+"ui-jx-home-")
			Expect(err).ShouldNot(HaveOccurred())

			_ = os.Setenv("JX_HOME", jxHome)
			utils.LogInfo(fmt.Sprintf("Using '%s' as JX_HOME", jxHome))

			By("setting the GitHub token")
			helpers.SetGitHubToken()

			By("setting up a GitHub client")
			ctx = context.Background()
			gitHubClient = t.GitHubClient()

			By("parsing the gitops dev repo information")
			gitInfo, err = gits.ParseGitURL(t.GitOpsDevRepo())
			Expect(err).ShouldNot(HaveOccurred())

			By("ensuring UI is not installed", func() {
				pr, err := helpers.GetPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Add %s %s", uiAppName, uiAppVersion))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(pr).Should(BeNil())
			})

			By("install UI via 'jx add app'", func() {
				By("installing the app")
				args := []string{"add", "app", uiAppName, "--version", uiAppVersion, "--repository=https://charts.cloudbees.com/cjxd/cloudbees"}
				t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)

				pr, err := helpers.GetPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Add %s %s", uiAppName, uiAppVersion))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(pr).ShouldNot(BeNil())
				Expect(*pr.State).Should(Equal("open"))

				// Wait for the pipeline to be started before merging to avoid pipelinerun non unique name bug
				// TODO: can be removed when builder includes https://github.com/jenkins-x/jx/pull/6322
				time.Sleep(30 * time.Second)

				By("merging the install app PR")
				results, _, err := gitHubClient.PullRequests.Merge(ctx, gitInfo.Organisation, gitInfo.Name, *pr.Number, "PR merge", nil)
				Expect(pr).ShouldNot(BeNil())
				Expect(*results.Merged).Should(BeTrue())

				By("waiting for the build to complete")
				jobName := fmt.Sprintf("%s/%s/master #%s", gitInfo.Organisation, gitInfo.Name, t.NextBuildNumber(gitInfo))
				t.TailBuildLog(jobName, helpers.TimeoutBuildCompletes)
			})

			By("ensure UI is installed", func() {
				args := []string{"get", "app", uiAppName}
				t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)
			})

			By("Accessing the UI", func() {
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
		}
	})

	AfterEach(func() {
		if runner.JxUiUrl() == "" {
			return
		}
		By("uninstalling the UI app", func() {
			if runner.JxUiUrl() == "" {
				return
			}
			args := []string{"delete", "app", uiAppName}
			t.ExpectJxExecution(t.WorkDir, timeoutAppTests, 0, args...)

			pr, err := helpers.GetPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, fmt.Sprintf("Delete %s", uiAppName))
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

		_ = os.RemoveAll(jxHome)
	})

	return Context("UI", func() {
		applicationName := ""
		JustBeforeEach(func() {
			By("Creating a new project from quickstart", func() {
				applicationName = createQuickstart("node-http")
			})
		})

		JustAfterEach(func() {
			By("Cleaning up quickstart", func() {
				cleanupQuickstart(applicationName)
			})
		})

		It("Runs smoke tests", func() {
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
			command.Env = append(command.Env, fmt.Sprintf("APPLICATION_NAME=%s", applicationName))
			command.Stderr = GinkgoWriter
			command.Stdout = GinkgoWriter
			err = command.Run()
			if err != nil {
				fmt.Println("Smoke tests failed")
			}
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
}

func createQuickstart(quickstartName string) string {
	qsNameParts := strings.Split(quickstartName, "-")
	qsAbbr := ""
	for s := range qsNameParts {
		qsAbbr = qsAbbr + qsNameParts[s][:1]

	}
	applicationName := helpers.TempDirPrefix + qsAbbr + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10)
	T := helpers.TestOptions{
		ApplicationName: applicationName,
		WorkDir:         helpers.WorkDir,
	}

	args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", applicationName, "-f", quickstartName, "--git-username", os.Getenv("GH_USERNAME")}

	gitProviderUrl, err := T.GitProviderURL()
	Expect(err).NotTo(HaveOccurred())
	if gitProviderUrl != "" {
		utils.LogInfof("Using Git provider URL %s\n", gitProviderUrl)
		args = append(args, "--git-provider-url", gitProviderUrl)
	}
	argsStr := strings.Join(args, " ")
	By(fmt.Sprintf("calling jx %s", argsStr), func() {
		T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
	})

	owner := T.GetGitOrganisation()
	jobName := owner + "/" + applicationName + "/master"

	//FIXME Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
	time.Sleep(30 * time.Second)
	By(fmt.Sprintf("waiting for the first release of %s", applicationName), func() {
		T.ThereShouldBeAJobThatCompletesSuccessfully(jobName, helpers.TimeoutBuildCompletes)
		T.TheApplicationIsRunningInStaging(200)
	})

	return applicationName
}

func cleanupQuickstart(applicationName string) {
	var T helpers.TestOptions

	if T.DeleteApplications() {
		args := []string{"delete", "application", "-b", applicationName}
		argsStr := strings.Join(args, " ")
		By(fmt.Sprintf("calling %s to delete the application", argsStr), func() {
			T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
		})
	}

	if T.DeleteRepos() {
		args := []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", applicationName}
		argsStr := strings.Join(args, " ")

		By(fmt.Sprintf("calling %s to delete the repository", argsStr), func() {
			T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
		})
	}
}
