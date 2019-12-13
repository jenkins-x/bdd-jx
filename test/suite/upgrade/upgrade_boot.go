package upgrade

import (
	"context"
	"fmt"
	"github.com/google/go-github/v28/github"
	"github.com/jenkins-x/bdd-jx/test/helpers"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/gits"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
)

type testCaseUpgradeBoot struct {
	helpers.TestOptions
	*runner.JxRunner
	client    kubernetes.Interface
	namespace string
}

func newTestCaseUpgradeBoot(cwd string, factory cmd.Factory) (*testCaseUpgradeBoot, error) {
	client, ns, err := factory.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &testCaseUpgradeBoot{
		JxRunner:  runner.New(cwd, nil, 0),
		client:    client,
		namespace: ns,
	}, nil
}

func (t *testCaseUpgradeBoot) upgrade() {
	allargs := []string{"upgrade", "boot", "-b"}
	upgradeVersionRef := os.Getenv("JX_UPGRADE_VERSION_REF")
	if upgradeVersionRef != "" {
		utils.LogInfo(fmt.Sprintf("Using upgrade ref: %s", upgradeVersionRef))
		allargs = append(allargs, fmt.Sprintf("--upgrade-version-stream-ref=%s", upgradeVersionRef))
	}
	t.Run(allargs...)
}
func (t *testCaseUpgradeBoot) overwriteJxBinary() {
	// TODO: We should get this working with jx upgrade cli
	jxBinDir := os.Getenv("JX_BIN_DIR")
	Expect(jxBinDir).To(BeADirectory())
	jxUpgradeBinDir := os.Getenv("JX_UPGRADE_BIN_DIR")
	Expect(jxUpgradeBinDir).To(BeADirectory())
	err := os.Remove(filepath.Join(jxBinDir, "jx"))
	Expect(err).NotTo(HaveOccurred())
	// Copy over the new binary
	err = os.Rename(filepath.Join(jxUpgradeBinDir, "jx"), filepath.Join(jxBinDir, "jx"))
	Expect(err).NotTo(HaveOccurred())
}

var _ = Describe("upgrade boot", func() {
	var (
		test         *testCaseUpgradeBoot
		jxHome       string
		ctx          context.Context
		gitHubClient *github.Client
		gitInfo      *gits.GitRepository
		err          error
	)

	BeforeEach(func() {
		test, err = newTestCaseUpgradeBoot(helpers.WorkDir, cmd.NewFactory())
		Expect(err).NotTo(HaveOccurred())
		Expect(test).NotTo(BeNil())
		By("setting a temporary JX_HOME directory")
		jxHome, err = ioutil.TempDir("", helpers.TempDirPrefix+"upgrade-boot-home-")
		Expect(err).ShouldNot(HaveOccurred())
		_ = os.Setenv("JX_HOME", jxHome)
		utils.LogInfo(fmt.Sprintf("Using '%s' as JX_HOME", jxHome))
	})

	BeforeEach(func() {
		By("setting the GitHub token")
		test.SetGitHubToken()
	})

	BeforeEach(func() {
		By("setting up a GitHub client")
		ctx = context.Background()
		gitHubClient = test.GitHubClient()
	})

	BeforeEach(func() {
		By("parsing the gitops dev repo information")
		gitInfo, err = gits.ParseGitURL(test.GitOpsDevRepo())
		Expect(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = os.RemoveAll(jxHome)
	})

	Describe("Given valid parameters", func() {
		Context("when running upgrade platform", func() {
			It("updates the platform to the given version", func() {
				if os.Getenv("JX_UPGRADE_BIN_DIR") != "" {
					test.overwriteJxBinary()
				} else {
					utils.LogInfo("JX_UPGRADE_BIN_DIR was not set so not upgrading using existing jx binary")
				}
				test.upgrade()
				pr, err := test.GetPullRequestWithTitle(gitHubClient, ctx, gitInfo.Organisation, gitInfo.Name, "feat(config): upgrade configuration")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(pr).ShouldNot(BeNil())
				Expect(*pr.State).Should(Equal("open"))

				By("merging the upgrade PR")
				results, _, err := gitHubClient.PullRequests.Merge(ctx, gitInfo.Organisation, gitInfo.Name, *pr.Number, "PR merge", nil)
				Expect(pr).ShouldNot(BeNil())
				Expect(*results.Merged).Should(BeTrue())

				By("waiting for the build to complete")
				jobName := fmt.Sprintf("%s/%s/master", gitInfo.Organisation, gitInfo.Name)
				By(fmt.Sprintf("checking that job %s completes successfully", jobName), func() {
					test.ThereShouldBeAJobThatCompletesSuccessfully(jobName, helpers.TimeoutBuildCompletes)
				})
			})
		})

	})

})
