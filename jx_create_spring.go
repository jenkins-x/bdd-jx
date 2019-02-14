package bdd_jx

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/util"
	"os/exec"
	"strconv"
	"strings"
	"time"

	cmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("create spring\n", func() {
	var T Test

	BeforeEach(func() {
		T = Test{
			ApplicationName: TempDirPrefix + "spring-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir:         WorkDir,
			Factory:         cmd.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given valid parameters", func() {
		Context("when running jx create spring", func() {
			It("creates a spring application and promotes it to staging\n", func() {
				c := "jx"
				args := []string{"create", "spring", "-b", "--org", T.GetGitOrganisation(), "--artifact", T.ApplicationName, "--name", T.ApplicationName, "-d", "web", "-d", "actuator"}

				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				if gitProviderUrl != "" {
					utils.LogInfof("Using Git provider URL %s\n", gitProviderUrl)
					args = append(args, "--git-provider-url", gitProviderUrl)
				}
				utils.LogInfof("about to run command: %s\n", util.ColorInfo(fmt.Sprintf("%s %s", c, strings.Join(args, " "))))

				command := exec.Command(c, args...)
				command.Dir = T.WorkDir
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred())
				session.Wait(1 * time.Hour)
				Eventually(session).Should(gexec.Exit(0))

				if T.WaitForFirstRelease() {
					T.TheApplicationShouldBeBuiltAndPromotedViaCICD(404)
				}

				if T.TestPullRequest() {
					By("perform a pull request on the source and assert that a preview environment is created")
					T.CreatePullRequestAndGetPreviewEnvironment(404)
				}

				if T.DeleteApplications() {
					By("deletes the application")
					args = []string{"delete", "application", "-b", T.ApplicationName}
					command = exec.Command(c, args...)
					command.Dir = T.WorkDir
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}

				if T.DeleteRepos() {
					By("deletes the repo")
					args = []string{"delete", "repo", "-b", "--github", "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
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
})
