package bdd_jx

import (
	"os/exec"
	"strconv"
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
			AppName: TempDirPrefix + "create-spring-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir: WorkDir,
			Factory: cmd.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given valid parameters", func() {
		Context("when running jx create spring", func() {
			It("creates a spring application and promotes it to staging\n", func() {
				c := "jx"
				args := []string{"create", "spring", "-b", "--org", T.GetGitOrganisation(), "--artifact", T.AppName, "--name", T.AppName}
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
})
