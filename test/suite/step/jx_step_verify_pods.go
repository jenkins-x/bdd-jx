package step

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jenkins-x/bdd-jx/test/helpers"
	"github.com/jenkins-x/bdd-jx/test/utils"
	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("verify pods", func() {

	var T StepTestOptions

	BeforeEach(func() {
		T = StepTestOptions{
			helpers.TestOptions{
				ApplicationName: helpers.TempDirPrefix + "verify-pods-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
				WorkDir:         helpers.WorkDir,
			},
		}
	})

	Describe("Verify there are no failed pods", func() {
		Context("by running jx step verify pod", func() {
			It("should exit 0 or contain the word Failed", func() {

				args := []string{"step", "verify", "pod", "ready"}
				argsStr := strings.Join(args, " ")
				var out string
				By(fmt.Sprintf("calling jx %s", argsStr), func() {
					r := runner.New(T.WorkDir, &helpers.TimeoutCmdLine, 0)
					var err error
					out, err = r.RunWithOutput(args...)
					utils.ExpectNoError(err)
				})
				Expect(out).ShouldNot(ContainSubstring("Failed"), "There are failed pods")
			})
		})
	})
})
