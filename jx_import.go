package bdd_jx

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jenkins-x/bdd-jx/utils"
	cmd "github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	git "gopkg.in/src-d/go-git.v4"
)

var _ = Describe("import\n", func() {

	var T Test

	BeforeEach(func() {
		T = Test{
			AppName: TempDirPrefix + "import-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir: WorkDir,
			Factory: cmd.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given valid parameters", func() {
		Context("when running import", func() {
			It("creates an app from the specified folder and promotes it to staging\n", func() {
				dest_dir := T.WorkDir + "/" + T.AppName

				_, err := git.PlainClone(dest_dir, false, &git.CloneOptions{
					URL:      "https://github.com/jenkins-x-quickstarts/spring-boot-watch-pipeline-activity.git",
					Progress: GinkgoWriter,
				})
				Expect(err).NotTo(HaveOccurred())
				os.RemoveAll(dest_dir + "/.git")
				Expect(dest_dir + "/.git").ToNot(BeADirectory())
				err = utils.ReplaceElement(filepath.Join(dest_dir, "pom.xml"), "artifactId", T.AppName, 1)
				Expect(err).NotTo(HaveOccurred())

				c := "jx"
				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				args := []string{"import", dest_dir, "-b", "--org", T.GetGitOrganisation(), "--git-provider-url", gitProviderUrl}
				command := exec.Command(c, args...)
				command.Dir = dest_dir
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred())
				session.Wait(1 * time.Hour)
				Eventually(session).Should(gexec.Exit(0))
				T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)

				if T.DeleteApps() {
					By("deletes the app")
					fullAppName := T.GetGitOrganisation() + "/" + T.AppName
					args = []string{"delete", "app", "-b", fullAppName}
					command = exec.Command(c, args...)
					command.Dir = dest_dir
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}

				if T.DeleteRepos() {
					By("deletes the repo")
					args = []string{"delete", "repo", "-b", "-g", gitProviderUrl, "-o", T.GetGitOrganisation(), "-n", T.AppName}
					command = exec.Command(c, args...)
					command.Dir = dest_dir
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}
			})
		})
	})
})
