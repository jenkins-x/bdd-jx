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

var _ = Describe("quickstart golang-http\n", func() {

	const quickstartName = "golang-http"
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
