package apps

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	. "github.com/onsi/ginkgo"
)

var _ = AppTests()

func AppTests() []bool {
	var appsUnderTest string
	apps, set := os.LookupEnv("JX_BDD_INCLUDE_APPS")
	if set {
		appsUnderTest = apps
	} else {
		appsUnderTest = IncludeApps
	}

	includedAppList := strings.Split(strings.TrimSpace(appsUnderTest), ",")
	tests := make([]bool, len(includedAppList))
	for _, testAppName := range includedAppList {
		nameAndVersion := strings.Split(testAppName, ":")
		if len(nameAndVersion) == 2 {
			tests = append(tests, AddAppTest(nameAndVersion[0], nameAndVersion[1]))
		} else {
			tests = append(tests, AddAppTest(testAppName, ""))
		}
	}
	return tests
}

type AppTestOptions struct {
	helpers.TestOptions
}

func AddAppTest(testAppName string, version string) bool {
	return Describe("Apps Framework", func() {
		var T AppTestOptions

		BeforeEach(func() {
			T = AppTestOptions{
				helpers.TestOptions{
					ApplicationName: helpers.TempDirPrefix + testAppName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
					WorkDir:         helpers.WorkDir,
				},
			}
			T.GitProviderURL()
		})

		_ = T.AddAppTests(testAppName, version)
		_ = T.GetAppsTests(testAppName)
		_ = T.UpgradeAppTests(testAppName)
		_ = T.DeleteAppTests(testAppName)

	})
}

// AddAppTests add app tests
func (t *AppTestOptions) AddAppTests(testAppName string, version string) bool {
	return Describe("Adding an app", func() {
		Context("by running jx add app "+testAppName, func() {
			It("should be added", func() {
				args := []string{"get", "app", testAppName}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("calling jx %s to check that the app does not exist before creation", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 1, args...)
				})

				args = []string{"add", "app", testAppName, "--repository", helpers.DefaultRepositoryURL}
				if version != "" {
					args = append(args, "--version", version)
				}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking that jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})

				args = []string{"get", "app", testAppName}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("calling jx %s to check that the app exists", args), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})
			})
		})
	})
}

func (t *AppTestOptions) GetAppsTests(testAppName string) bool {
	return Describe("Getting an app", func() {
		Context("by running jx get apps "+testAppName, func() {
			It("should display the correct output", func() {
				args := []string{"get", "app", testAppName}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})

				args = []string{"get", "app", testAppName, "-o", "yaml"}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})

				args = []string{"get", "app", testAppName, "-o", "json"}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})
			})
		})
	})
}

func (t *AppTestOptions) UpgradeAppTests(testAppName string) bool {
	return Describe("Upgrading an app", func() {
		Context("by running jx upgrade app "+testAppName, func() {
			It("should be upgraded", func() {
				args := []string{"get", "app", testAppName}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})
				args = []string{"upgrade", "app", testAppName}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})
			})
		})
	})
}

// DeleteAppTests delete app tests
func (t *AppTestOptions) DeleteAppTests(testAppName string) bool {
	return Describe("Deleting an app", func() {
		Context("by running jx delete app "+testAppName, func() {
			It("should be deleted", func() {
				args := []string{"get", "app", testAppName}
				argsStr := strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {
					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})
				args = []string{"delete", "app", testAppName}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {

					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 0, args...)
				})

				args = []string{"get", "app", testAppName}
				argsStr = strings.Join(args, " ")
				By(fmt.Sprintf("checking jx %s exits with signal 0", argsStr), func() {

					t.ExpectJxExecution(t.WorkDir, TimeoutAppTests, 1, args...)
				})
			})
		})
	})
}
