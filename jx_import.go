package bdd_jx

import (
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/jx/cmd/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"path/filepath"
	"strconv"
)

var _ = Describe("import\n", func() {

	var T Test

	BeforeEach(func() {
		T = Test{
			ApplicationName: TempDirPrefix + "import-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir:         WorkDir,
			Factory:         clients.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given valid parameters", func() {
		Context("when running import", func() {
			It("creates an application from the specified folder and promotes it to staging\n", func() {
				dest_dir := T.WorkDir + "/" + T.ApplicationName

				_, err := git.PlainClone(dest_dir, false, &git.CloneOptions{
					URL:      "https://github.com/jenkins-x-quickstarts/spring-boot-watch-pipeline-activity.git",
					Progress: GinkgoWriter,
				})
				Expect(err).NotTo(HaveOccurred())
				os.RemoveAll(dest_dir + "/.git")
				Expect(dest_dir + "/.git").ToNot(BeADirectory())
				err = utils.ReplaceElement(filepath.Join(dest_dir, "pom.xml"), "artifactId", T.ApplicationName, 1)
				Expect(err).NotTo(HaveOccurred())

				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				args := []string{"import", dest_dir, "-b", "--org", T.GetGitOrganisation(), "--git-provider-url", gitProviderUrl}
				T.ExpectJxExecution(T.WorkDir, TimeoutSessionWait, 0, args...)
				T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)

				if T.DeleteApplications() {
					By("deletes the application")
					args = []string{"delete", "application", "-b", T.ApplicationName}
					T.ExpectJxExecution(T.WorkDir, TimeoutSessionWait, 0, args...)
				}

				if T.DeleteRepos() {
					By("deletes the repo")
					args = []string{"delete", "repo", "-b", "-g", gitProviderUrl, "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
					T.ExpectJxExecution(T.WorkDir, TimeoutSessionWait, 0, args...)
				}
			})
		})
	})
})
