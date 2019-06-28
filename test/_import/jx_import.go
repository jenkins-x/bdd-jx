package _import

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	"github.com/jenkins-x/bdd-jx/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	git "gopkg.in/src-d/go-git.v4"
)

type ImportTestOptions struct {
	helpers.TestOptions
}

var _ = Describe("Import Application", func() {

	var T ImportTestOptions

	BeforeEach(func() {
		T = ImportTestOptions{
			helpers.TestOptions{
				ApplicationName: helpers.TempDirPrefix + "import-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         helpers.WorkDir,
			},
		}
		T.GitProviderURL()
	})

	Describe("Importing an application", func() {
		Context("by running jx import", func() {
			It("creates an application from the specified folder and promotes it to staging", func() {
				destDir := T.WorkDir + "/" + T.ApplicationName
				url := "https://github.com/jenkins-x-quickstarts/spring-boot-watch-pipeline-activity.git"

				By(fmt.Sprintf("calling git clone %s", url), func() {
					_, err := git.PlainClone(destDir, false, &git.CloneOptions{
						URL:      url,
						Progress: GinkgoWriter,
					})
					Expect(err).NotTo(HaveOccurred())
				})

				By("removing the .git directory", func() {
					err := os.RemoveAll(destDir + "/.git")
					utils.ExpectNoError(err)
					Expect(destDir + "/.git").ToNot(BeADirectory())
				})

				By("updating the pom.xml to have the correct application name", func() {
					err := utils.ReplaceElement(filepath.Join(destDir, "pom.xml"), "artifactId", T.ApplicationName, 1)
					Expect(err).NotTo(HaveOccurred())
				})

				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				args := []string{"import", destDir, "-b", "--org", T.GetGitOrganisation(), "--git-provider-url", gitProviderUrl}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("running jx %s", argsStr), func() {
					T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
				})

				T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)

				if T.DeleteApplications() {
					args = []string{"delete", "application", "-b", T.ApplicationName}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("deleting the application by calling jx %s", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
					})

				}

				if T.DeleteRepos() {
					args = []string{"delete", "repo", "-b", "-g", gitProviderUrl, "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
					argsStr := strings.Join(args, " ")
					By(fmt.Sprintf("deleting the repo by calling jx %s", argsStr), func() {
						T.ExpectJxExecution(T.WorkDir, helpers.TimeoutSessionWait, 0, args...)
					})
				}
			})
		})
	})
})
