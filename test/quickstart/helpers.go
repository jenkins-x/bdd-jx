package quickstart

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/jx/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = AllQuickstartsTest()

var (
	IncludeQuickstarts = os.Getenv("JX_BDD_QUICKSTARTS")
)

//CreateQuickstartTest creates a quickstart test for the given quickstart
func CreateQuickstartTest(quickstartName string) bool {
	return createQuickstartTests(quickstartName, false)
}

// AddAppTests Creates a jx add app test
func AllQuickstartsTest() []bool {
	if IncludeQuickstarts != "" {
		includedQuickstartList := strings.Split(strings.TrimSpace(IncludeQuickstarts), ",")
		tests := make([]bool, len(includedQuickstartList))
		for _, testQuickstartName := range includedQuickstartList {
			tests = append(tests, CreateBatchQuickstartsTests(testQuickstartName))
		}
		return tests
	} else {
		return make([]bool, 0)
	}
}

//CreateBatchQuickstartsTests creates a batch quickstart test for the given quickstart
func CreateBatchQuickstartsTests(quickstartName string) bool {
	return createQuickstartTests(quickstartName, true)
}

// CreateQuickstartTest Creates quickstart tests.  If batch == true, add 'batch' to the test spec
func createQuickstartTests(quickstartName string, batch bool) bool {
	description := ""
	if batch {
		description = "[batch] "
	}
	return Describe(description+"quickstart "+quickstartName+"\n", func() {
		var T helpers.TestOptions

		BeforeEach(func() {
			qsNameParts := strings.Split(quickstartName, "-")
			qsAbbr := ""
			for s := range qsNameParts {
				qsAbbr = qsAbbr + qsNameParts[s][:1]

			}
			applicationName := helpers.TempDirPrefix + qsAbbr + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10)
			T = helpers.TestOptions{
				ApplicationName: applicationName,
				WorkDir:         helpers.WorkDir,
			}
			T.GitProviderURL()

			utils.LogInfof("Creating application %s in dir %s\n", util.ColorInfo(applicationName), util.ColorInfo(helpers.WorkDir))
		})

		Describe("Create a quickstart", func() {
			Context(fmt.Sprintf("by running jx create quickstart %s", quickstartName), func() {
				It("creates a new source repository and promotes it to staging", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", quickstartName}

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

					applicationName := T.GetApplicationName()
					owner := T.GetGitOrganisation()
					jobName := owner + "/" + applicationName + "/master"

					if T.WaitForFirstRelease() {
						//FIXME Need to wait a little here to ensure that the build has started before asking for the log as the jx create quickstart command returns slightly before the build log is available
						time.Sleep(30 * time.Second)
						By(fmt.Sprintf("waiting for the first release of %s", applicationName), func() {
							T.ThereShouldBeAJobThatCompletesSuccessfully(jobName, helpers.TimeoutBuildCompletes)
							T.TheApplicationIsRunningInStaging(200)
						})

						if T.TestPullRequest() {
							By("performing a pull request on the source and asserting that a preview environment is created", func() {
								T.CreatePullRequestAndGetPreviewEnvironment(200)
							})
						}
					} else {
						By(fmt.Sprintf("waiting for the first successful build of master of %s", applicationName), func() {
							T.ThereShouldBeAJobThatCompletesSuccessfully(jobName, helpers.TimeoutBuildCompletes)
						})
					}

					if T.DeleteApplications() {
						args = []string{"delete", "application", "-b", T.ApplicationName}
						argsStr := strings.Join(args, " ")
						By(fmt.Sprintf("calling %s to delete the application", argsStr), func() {
							T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
						})
					}

					if T.DeleteRepos() {
						args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
						argsStr = strings.Join(args, " ")

						By(fmt.Sprintf("calling %s to delete the repository", os.Args), func() {
							T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
						})
					}
				})
			})
		})
		Describe("Create a quickstart with invalid parameters", func() {
			Context("when -p param (project name) is missing", func() {
				It("exits with signal 1", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-f", quickstartName}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("calling jx %s", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 1, args...)
					})
				})
			})
			Context("when -f param (filter) does not match any quickstart", func() {
				It("exits with signal 1", func() {
					args := []string{"create", "quickstart", "-b", "--org", T.GetGitOrganisation(), "-p", T.ApplicationName, "-f", "the_derek_zoolander_app_for_being_really_really_good_looking"}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("calling jx %s", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 1, args...)
					})
				})
			})
		})
	})
}
