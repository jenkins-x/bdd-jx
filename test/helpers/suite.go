package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jenkins-x/jx/pkg/client/clientset/versioned"
	cmd "github.com/jenkins-x/jx/pkg/cmd/clients"
	"github.com/jenkins-x/jx/pkg/kube"
	"github.com/onsi/ginkgo/config"
	"k8s.io/client-go/kubernetes"

	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	"github.com/pkg/errors"

	gr "github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunWithReporters(t *testing.T, suiteId string) {
	reportsDir := os.Getenv("REPORTS_DIR")
	if reportsDir == "" {
		reportsDir = filepath.Join("../", "build", "reports")
	}
	err := os.MkdirAll(reportsDir, 0700)
	if err != nil {
		t.Errorf("cannot create %s because %v", reportsDir, err)
	}
	reporters := make([]Reporter, 0)

	slowSpecThresholdStr := os.Getenv("SLOW_SPEC_THRESHOLD")
	if slowSpecThresholdStr == "" {
		slowSpecThresholdStr = "50000"
		_ = os.Setenv("SLOW_SPEC_THRESHOLD", slowSpecThresholdStr)

	}
	slowSpecThreshold, err := strconv.ParseFloat(slowSpecThresholdStr, 64)
	if err != nil {
		panic(err.Error())
	}
	config.DefaultReporterConfig.SlowSpecThreshold = slowSpecThreshold
	config.DefaultReporterConfig.Verbose = testing.Verbose()
	reporters = append(reporters, gr.NewJUnitReporter(filepath.Join(reportsDir, fmt.Sprintf("%s.junit.xml", suiteId))))
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("Jenkins X E2E tests: %s", suiteId), reporters)
}

var BeforeSuiteCallback = func() {
	err := ensureConfiguration()
	utils.ExpectNoError(err)
	WorkDir, err := ioutil.TempDir("", TempDirPrefix)
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(WorkDir, 0760)
	Expect(err).NotTo(HaveOccurred())
	Expect(WorkDir).To(BeADirectory())
	AssignWorkDirValue(WorkDir)
}

var SynchronizedAfterSuiteCallback = func() {
	// Cleanup workdir as usual
	cleanFlag := os.Getenv("JX_DISABLE_CLEAN_DIR")
	if strings.ToLower(cleanFlag) != "true" {
		os.RemoveAll(WorkDir)
		Expect(WorkDir).ToNot(BeADirectory())
	}
}

func ensureConfiguration() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.WithStack(err)
	}

	_, found := os.LookupEnv("BDD_JX")
	if !found {
		_ = os.Setenv("BDD_JX", runner.Jx)
	}

	r := runner.New(cwd, &TimeoutSessionWait, 0)
	version, err := r.RunWithOutput("--version")
	if err != nil {
		return errors.WithStack(err)
	}
	factory := cmd.NewFactory()
	kubeClient, ns, err := factory.CreateKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create kubeClient")
	}
	jxClient, _, err := factory.CreateJXClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to create jxClient")
	}

	gitOrganisation := os.Getenv("GIT_ORGANISATION")
	if gitOrganisation == "" {
		gitOrganisation, err = findDefaultOrganisation(kubeClient, jxClient, ns)
		if err != nil {
			return errors.Wrapf(errors.WithStack(err), "failed to find gitOrganisation in namespace %s", ns)
		}
		if gitOrganisation == "" {
			gitOrganisation = "jenkins-x-tests"
		}
		_ = os.Setenv("GIT_ORGANISATION", gitOrganisation)
	}
	gitProviderUrl := os.Getenv("GIT_PROVIDER_URL")
	if gitProviderUrl == "" {
		gitProviderUrl = "https://github.com"
		_ = os.Setenv("GIT_PROVIDER_URL", gitProviderUrl)
	}
	gitKind := os.Getenv("GIT_KIND")
	if gitKind == "" {
		gitKind = "github"
		os.Setenv("GIT_KIND", gitKind)
	}
	disableDeleteAppStr := os.Getenv("JX_DISABLE_DELETE_APP")
	disableDeleteApp := "is set. Apps created in the test run will NOT be deleted"
	if disableDeleteAppStr == "true" || disableDeleteAppStr == "1" || disableDeleteAppStr == "on" {
		disableDeleteApp = "is not set. If you would like to disable the automatic deletion of apps created by the tests set this variable to TRUE."
	}
	disableDeleteRepoStr := os.Getenv("JX_DISABLE_DELETE_REPO")
	disableDeleteRepo := "is set. Repos created in the test run will NOT be deleted"
	if disableDeleteRepoStr == "true" || disableDeleteRepoStr == "1" || disableDeleteRepoStr == "on" {
		disableDeleteRepo = "is not set. If you would like to disable the automatic deletion of repos created by the tests set this variable to TRUE."
	}
	disableWaitForFirstReleaseStr := os.Getenv("JX_DISABLE_WAIT_FOR_FIRST_RELEASE")
	disableWaitForFirstRelease := "is set. Will not wait for build to be promoted to staging"
	if disableWaitForFirstReleaseStr == "true" || disableWaitForFirstReleaseStr == "1" || disableWaitForFirstReleaseStr == "on" {
		disableWaitForFirstRelease = "is not set. If you would like to disable waiting for the build to be promoted to staging set this variable to TRUE"
	}
	enableChatOpsTestLogStr := "is not set. ChatOps tests will not be run as part of quickstart tests. If you would like to run those tests, set this variable to TRUE"
	if EnableChatOpsTests == "true" {
		enableChatOpsTestLogStr = "is set. ChatOps tests will be run as part of quickstart tests"
	}
	includeAppsStr := os.Getenv("JX_BDD_INCLUDE_APPS")
	includeApps := "is not set"
	if includeAppsStr != "" {
		includeApps = fmt.Sprintf("is set to %s", includeAppsStr)
	}
	bddTimeoutBuildCompletes := os.Getenv("BDD_TIMEOUT_BUILD_COMPLETES")
	if bddTimeoutBuildCompletes == "" {
		_ = os.Setenv("BDD_TIMEOUT_BUILD_COMPLETES", "60")
	}
	bddTimeoutBuildRunningInStaging := os.Getenv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING")
	if bddTimeoutBuildRunningInStaging == "" {
		_ = os.Setenv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING", "60")
	}
	bddTimeoutURLReturns := os.Getenv("BDD_TIMEOUT_URL_RETURNS")
	if bddTimeoutURLReturns == "" {
		_ = os.Setenv("BDD_TIMEOUT_URL_RETURNS", "5")
	}
	bddTimeoutCmdLine := os.Getenv("BDD_TIMEOUT_CMD_LINE")
	if bddTimeoutCmdLine == "" {
		_ = os.Setenv("BDD_TIMEOUT_CMD_LINE", "1")
	}
	bddTimeoutAppTests := os.Getenv("BDD_TIMEOUT_APP_TESTS")
	if bddTimeoutAppTests == "" {
		_ = os.Setenv("BDD_TIMEOUT_APP_TESTS", "60")
	}
	bddTimeoutSessionWait := os.Getenv("BDD_TIMEOUT_SESSION_WAIT")
	if bddTimeoutSessionWait == "" {
		_ = os.Setenv("BDD_TIMEOUT_SESSION_WAIT", "60")
	}
	bddTimeoutDevpod := os.Getenv("BDD_TIMEOUT_DEVPOD")
	if bddTimeoutDevpod == "" {
		_ = os.Setenv("BDD_TIMEOUT_DEVPOD", "15")
	}

	gheUser := os.Getenv("GHE_USER")
	if gheUser == "" {
		gheUser = "dev1"
		_ = os.Setenv("GHE_USER", gheUser)
	}
	gheProviderUrl := os.Getenv("GHE_PROVIDER_URL")
	if gheProviderUrl == "" {
		gheProviderUrl = "https://github.beescloud.com"
		_ = os.Setenv("GHE_PROVIDER_URL", gheProviderUrl)
	}

	utils.LogInfof("BDD_JX:                                             %s\n", os.Getenv("BDD_JX"))
	utils.LogInfof("jx version:                                         %s\n", version)
	utils.LogInfof("GIT_ORGANISATION:                                   %s\n", gitOrganisation)
	utils.LogInfof("GIT_PROVIDER_URL:                                   %s\n", gitProviderUrl)
	utils.LogInfof("GIT_KIND:                                           %s\n", gitKind)
	utils.LogInfof("JX_DISABLE_DELETE_APP:                              %s\n", disableDeleteApp)
	utils.LogInfof("JX_DISABLE_DELETE_REPO:                             %s\n", disableDeleteRepo)
	utils.LogInfof("JX_DISABLE_WAIT_FOR_FIRST_RELEASE:                  %s\n", disableWaitForFirstRelease)
	utils.LogInfof("JX_ENABLE_TEST_CHATOPS_COMMANDS:                    %s\n", enableChatOpsTestLogStr)
	utils.LogInfof("JX_BDD_INCLUDE_APPS:                                %s\n", includeApps)
	utils.LogInfof("BDD_TIMEOUT_BUILD_COMPLETES timeout value:          %s\n", os.Getenv("BDD_TIMEOUT_BUILD_COMPLETES"))
	utils.LogInfof("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING timeout value: %s\n", os.Getenv("BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING"))
	utils.LogInfof("BDD_TIMEOUT_URL_RETURNS timeout value:              %s\n", os.Getenv("BDD_TIMEOUT_URL_RETURNS"))
	utils.LogInfof("BDD_TIMEOUT_CMD_LINE timeout value:                 %s\n", os.Getenv("BDD_TIMEOUT_CMD_LINE"))
	utils.LogInfof("BDD_TIMEOUT_APP_TESTS timeout value:                %s\n", os.Getenv("BDD_TIMEOUT_APP_TESTS"))
	utils.LogInfof("BDD_TIMEOUT_SESSION_WAIT timeout value:             %s\n", os.Getenv("BDD_TIMEOUT_SESSION_WAIT"))
	utils.LogInfof("BDD_TIMEOUT_DEVPOD timeout value:             	   %s\n", os.Getenv("BDD_TIMEOUT_DEVPOD"))
	utils.LogInfof("SLOW_SPEC_THRESHOLD:                                %s\n", os.Getenv("SLOW_SPEC_THRESHOLD"))
	utils.LogInfof("GHE_USER:                                           %s\n", os.Getenv("GHE_USER"))
	utils.LogInfof("GHE_TOKEN:                                          %s\n", os.Getenv("GHE_TOKEN"))
	utils.LogInfof("GHE_PROVIDER_URL:                                   %s\n", os.Getenv("GHE_PROVIDER_URL"))
	return nil
}

func findDefaultOrganisation(kubeClient kubernetes.Interface, jxClient versioned.Interface, ns string) (string, error) {
	// lets see if we have defined a team environment
	devEnv, err := kube.GetDevEnvironment(jxClient, ns)
	if err != nil {
		utils.LogInfof("failed to find the dev environment in namespace %s due to %s", ns, err.Error())
	}
	answer := ""
	if devEnv != nil {
		answer = devEnv.Spec.TeamSettings.Organisation
		if answer == "" {
			answer = devEnv.Spec.TeamSettings.EnvOrganisation
		}
		if answer == "" {
			answer = devEnv.Spec.TeamSettings.PipelineUsername
		}
		if answer != "" {
			return answer, nil
		}
	}
	utils.LogInfof("found organisation in namespace %s due to %s\n", ns, answer)

	// TODO load the user from the git secrets?
	return "", nil
}
