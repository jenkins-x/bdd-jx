package bdd_jx

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/utils"
	"github.com/jenkins-x/jx/pkg/jx/cmd"
	"github.com/jenkins-x/jx/pkg/util"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var _ = Describe("verify pods\n", func() {

	utils.LogInfof("About to verify pods")
	var T Test

	BeforeEach(func() {
		T = Test{
			AppName: TempDirPrefix + "verify-pods-" + strconv.FormatInt(GinkgoRandomSeed(), 10),
			WorkDir: WorkDir,
			Factory: cmd.NewFactory(),
		}
		T.GitProviderURL()
	})

	Describe("Given a completed test run", func() {
		Context("when running jx step verify pod", func() {
			It("there are no failed pods\n", func() {
				c := "jx"
				args := []string{"step", "verify", "pod"}

				utils.LogInfof("about to run command: %s\n", util.ColorInfo(fmt.Sprintf("%s %s", c, strings.Join(args, " "))))

				command := exec.Command(c, args...)
				command.Dir = T.WorkDir

				// fake the output stream to be checked later
				r, fakeStdout, _ := os.Pipe()

				session, err := gexec.Start(command, fakeStdout, GinkgoWriter)
				Ω(err).ShouldNot(HaveOccurred())
				session.Wait(1 * time.Minute)

				// check output
				fakeStdout.Close()
				outBytes, err := ioutil.ReadAll(r)
				r.Close()

				Ω(err).ShouldNot(HaveOccurred())

				Expect(string(outBytes)).ShouldNot(ContainSubstring("Failed"), "There are failed pods")

				Eventually(session).Should(gexec.Exit(0))

			})
		})
	})
	utils.LogInfof("Pods verified")
})
