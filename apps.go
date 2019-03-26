package bdd_jx

import (
	"github.com/jenkins-x/jx/pkg/jx/cmd/clients"
	. "github.com/onsi/ginkgo"
	"strconv"
	"strings"
)

func AppTests() []bool {
	if IncludeApps != "" {
		includedAppList := strings.Split(strings.TrimSpace(IncludeApps), ",")
		tests := make([]bool, len(includedAppList))
		for _, testAppName := range includedAppList {
			nameAndVersion := strings.Split(testAppName, ":")
			if len(nameAndVersion) == 2 {
				tests = append(tests, AppTest(nameAndVersion[0], nameAndVersion[1]))
			} else {
				tests = append(tests, AppTest(testAppName, ""))
			}
		}
		return tests
	} else {
		return nil
	}
}

func AppTest(testAppName string, version string) bool {
	return Describe("test app "+testAppName+"\n", func() {
		var T Test

		BeforeEach(func() {
			T = Test{
				ApplicationName: TempDirPrefix + testAppName + "-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         WorkDir,
				Factory:         clients.NewFactory(),
			}
			T.GitProviderURL()
		})

		_ = T.AddAppTests(testAppName, version)
		_ = T.DeleteAppTests(testAppName)

	})
}

// AddAppTests add app tests
func (t *Test) AddAppTests(testAppName string, version string) bool {
	return Describe("Given valid parameters", func() {
		Context("when running jx add app "+testAppName, func() {

			It("Ensure the app is added\n", func() {
				By("The App resource does not exist before creation\n")
				c := "jx"
				args := []string{"get", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 1, c, args...)
				By("Add app exits with signal 0\n")
				c = "jx"
				args = []string{"add", "app", testAppName, "--repository", DefaultRepositoryURL}
				if version != "" {
					args = append(args, "--version", version)
				}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("The App resource exists after creation\n")
				c = "jx"
				args = []string{"get", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
			})
		})
	})
}

// DeleteAppTests delete app tests
func (t *Test) DeleteAppTests(testAppName string) bool {
	return Describe("Given valid parameters", func() {
		Context("when running jx delete app "+testAppName, func() {
			It("Ensure it is deleted\n", func() {
				By("The App resource exists before deletion\n")
				c := "jx"
				args := []string{"get", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("Delete app exits with signal 0\n")
				c = "jx"
				args = []string{"delete", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 0, c, args...)
				By("The App resource was removed\n")
				c = "jx"
				args = []string{"get", "app", testAppName}
				t.ExpectCommandExecution(t.WorkDir, TimeoutAppTests, 1, c, args...)
			})
		})
	})
}