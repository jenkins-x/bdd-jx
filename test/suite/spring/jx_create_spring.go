package spring

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	"github.com/jenkins-x/bdd-jx/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	SkipManualPromotion = os.Getenv("JX_BDD_SKIP_MANUAL_PROMOTION")
)

var _ = Describe("create spring\n", func() {
	var T SpringTestOptions

	BeforeEach(func() {
		T = SpringTestOptions{

			helpers.TestOptions{
				ApplicationName: helpers.TempDirPrefix + "spring-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         helpers.WorkDir,
			},
		}
		T.GitProviderURL()
	})

	Describe("Given valid parameters", func() {
		Context("when running jx create spring", func() {
			It("creates a spring application and promotes it to staging\n", func() {
				args := []string{"create", "spring", "-b", "--org", T.GetGitOrganisation(), "--artifact", T.ApplicationName, "--name", T.ApplicationName, "-d", "web", "-d", "actuator"}

				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				if gitProviderUrl != "" {
					utils.LogInfof("Using Git provider URL %s", gitProviderUrl)
					args = append(args, "--git-provider-url", gitProviderUrl)
				}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("calling jx %s", argsStr), func() {
					T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
				})
				if T.WaitForFirstRelease() {
					By(fmt.Sprintf("waiting for the first release"), func() {
						T.TheApplicationShouldBeBuiltAndPromotedViaCICD(404)
					})
				}

				if T.TestPullRequest() {
					By("performing a pull request on the source and asserting that a preview environment is created", func() {
						T.CreatePullRequestAndGetPreviewEnvironment(404)
					})
				}

				if SkipManualPromotion == "" {
					args = []string{"promote", "--env", "production", "--version", "0.0.1", T.ApplicationName}
					By("manually promoting app to production environment", func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
						T.TheApplicationIsRunningInProduction(404)
					})
				}

				if T.DeleteApplications() {
					args = []string{"delete", "application", "-b", T.ApplicationName}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("calling jx %s to delete the application", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
					})
				}

				if T.DeleteRepos() {
					args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("calling jx %s to delete the git repository", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
					})
				}
			})
		})
	})
})
