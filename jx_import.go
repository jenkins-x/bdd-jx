package bdd_jx

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/auth"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

const (
	c = "jx"
)

var _ = Describe("import\n", func() {

	var T Test

	BeforeEach(func() {
		T = Test{
			ApplicationName: TempDirPrefix + "import-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir:         WorkDir,
			Factory:         cmd.NewFactory(),
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
				command := exec.Command(c, args...)
				command.Dir = dest_dir
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred())
				session.Wait(1 * time.Hour)
				Eventually(session).Should(gexec.Exit(0))
				T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)

				if T.DeleteApplications() {
					By("deletes the application")
					fullApplicationName := T.GetGitOrganisation() + "/" + T.ApplicationName
					args = []string{"delete", "application", "-b", fullApplicationName}
					command = exec.Command(c, args...)
					command.Dir = dest_dir
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}

				if T.DeleteRepos() {
					By("deletes the repo")
					args = []string{"delete", "repo", "-b", "-g", gitProviderUrl, "-o", T.GetGitOrganisation(), "-n", T.ApplicationName}
					command = exec.Command(c, args...)
					command.Dir = dest_dir
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}
			})
		})

		Context("when running import a specified URL", func() {
			It("clones a repository and imports an application from the specified url and promotes it to staging\n", func() {
				org := T.GetGitOrganisation()
				repoName := "golang-http-" + strconv.FormatInt(GinkgoRandomSeed(), 10)
				T.ApplicationName = repoName
				gitProviderUrl, err := T.GitProviderURL()
				Expect(err).NotTo(HaveOccurred())
				createForkedRepoOf(T, "https://github.com/jenkins-x-quickstarts/golang-http.git", gitProviderUrl, org, repoName)

				repoUrl := fmt.Sprintf("%s/%s/%s.git", gitProviderUrl, org, repoName)
				args := []string{"import", "--url", repoUrl, "-b", "--org", T.GetGitOrganisation(), "--git-provider-url", gitProviderUrl}
				command := exec.Command(c, args...)
				session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred())
				session.Wait(1 * time.Hour)
				Eventually(session).Should(gexec.Exit(0))
				T.TheApplicationShouldBeBuiltAndPromotedViaCICD(200)

				if T.DeleteApplications() {
					By("deletes the application")
					fullApplicationName := T.GetGitOrganisation() + "/" + repoName
					args = []string{"delete", "application", "-b", fullApplicationName}
					command = exec.Command(c, args...)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}

				if T.DeleteRepos() {
					By("deletes the repo")
					args = []string{"delete", "repo", "-b", "-g", gitProviderUrl, "-o", T.GetGitOrganisation(), "-n", repoName}
					command = exec.Command(c, args...)
					session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
					Ω(err).ShouldNot(HaveOccurred())
					session.Wait(1 * time.Hour)
					Eventually(session).Should(gexec.Exit(0))
				}
			})
		})
	})
})

func createForkedRepoOf(T Test, originUrl, gitProviderUrl, org, repoName string) {
	dest_dir := T.WorkDir + "/" + T.ApplicationName + "-url"
	// Ideally we'd just fork the upstream repo we want to import, but github only allows us to have one
	// fork of another repo (even if you rename it). This would cause problems running two BDD tests
	// concurrently. Instead, clone the upstream locally, create a _new_ repo and push the code to that.
	localRepo, err := git.PlainClone(dest_dir, false, &git.CloneOptions{
		URL:      originUrl,
		Progress: GinkgoWriter,
	})
	Expect(err).NotTo(HaveOccurred())

	// Create a github client
	gitAuth, err := getGitAuth(T, gitProviderUrl)
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gitAuth.ApiToken},
	)
	var client *github.Client
	if gitProviderUrl == "https://github.com" || gitProviderUrl == "" {
		client = github.NewClient(oauth2.NewClient(ctx, ts))
	} else {
		// assume github enterprise
		client, err = github.NewEnterpriseClient(gitProviderUrl+"/api/v3", gitProviderUrl+"/api/v3/upload", oauth2.NewClient(ctx, ts))
		Expect(err).NotTo(HaveOccurred())
	}

	// Create a new repo with a unique name on github
	githubRepo, _, err := client.Repositories.Create(ctx, org, &github.Repository{Name: &repoName})
	Expect(err).NotTo(HaveOccurred())
	utils.LogInfof("Created repo %s\n", *githubRepo.Name)

	// Delete repo after use
	remoteName := "newfork"
	_, err = localRepo.CreateRemote(&config.RemoteConfig{
		Name: remoteName,
		URLs: []string{*githubRepo.CloneURL},
	})
	Expect(err).NotTo(HaveOccurred())
	// Push our local clone to github
	err = localRepo.Push(&git.PushOptions{
		RemoteName: remoteName,
		Auth: &http.BasicAuth{
			Username: gitAuth.Username,
			Password: gitAuth.ApiToken,
		},
	})
	Expect(err).NotTo(HaveOccurred())

	// Removed the temporary clone
	_ = os.RemoveAll(dest_dir)
}

func getGitAuth(t Test, gitUrl string) (*auth.UserAuth, error) {
	service, err := t.Factory.CreateAuthConfigService("gitAuth.yaml")
	Expect(err).NotTo(HaveOccurred())

	authService, err := service.LoadConfig()
	Expect(err).NotTo(HaveOccurred())
	auths := authService.FindUserAuths(gitUrl)
	if len(auths) == 0 {
		return nil, fmt.Errorf("no git credentials for %s", gitUrl)
	} else if len(auths) > 1 {
		authNames := make([]string, len(auths))
		for i, auth := range auths {
			authNames[i] = auth.Username
		}
		utils.LogInfof("%v sets of credentials for %s: %v. Using first (%s)", len(auths), gitUrl, authNames, auths[0].Username)
	}
	return auths[0], nil
}
