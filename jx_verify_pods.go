package bdd_jx

import (
	"github.com/jenkins-x/bdd-jx/runner"
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/jx/cmd/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"time"
)

var _ = Describe("verify pods\n", func() {

	utils.LogInfof("About to verify pods")
	var T Test

	BeforeEach(func() {
		T = Test{
			ApplicationName: TempDirPrefix + "verify-pods-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir:         WorkDir,
			Factory:         clients.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given a completed test run", func() {
		Context("when running jx step verify pod", func() {
			It("there are no failed pods\n", func() {
				args := []string{"step", "verify", "pod", "ready"}

				timeout := 1 * time.Minute
				r := runner.New(T.WorkDir, &timeout, 0)
				out := r.RunWithOutput(args...)

				Expect(out).ShouldNot(ContainSubstring("Failed"), "There are failed pods")
			})
		})
	})
	utils.LogInfof("Pods verified")
})
