package runner

import (
	"github.com/jenkins-x/bdd-jx/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const (
	jx = "jx"
)

var (
	// jxRunner session timeout
	TimeoutJxRunner = utils.GetTimeoutFromEnv("BDD_TIMEOUT_JX_RUNNER", 5)
	coverageOutputRegex = regexp.MustCompile(`(?m:(PASS|FAIL)\n\s*coverage: ([\d\.]*%) of statements in [\w\.\/]*)`)
)



// Runner runs a jx command
type JxRunner struct {
	cwd string
	timeout time.Duration
	exitCode int
}

// New creates a new jx command runnner
func New(cwd string, timeout* time.Duration, exitCode int) *JxRunner {
	if timeout == nil {
		timeout = &TimeoutJxRunner
	}
	return &JxRunner{
		cwd:      cwd,
		timeout:  *timeout,
		exitCode: exitCode,
	}
}

// Run runs a jx command
func (r *JxRunner) Run(args ...string) {
	r.run(GinkgoWriter, GinkgoWriter, args...)
}

func (r *JxRunner) run(out io.Writer, errOut io.Writer, args ...string) {
	command := exec.Command(jx, args...)
	command.Dir = r.cwd
	session, err := gexec.Start(command, out, errOut)
	utils.ExpectNoError(err)
	session.Wait(r.timeout)
	Eventually(session).Should(gexec.Exit(r.exitCode), "whilst running command %s %s", jx, strings.Join(args, " "))
}

// Run runs a jx command
func (r *JxRunner) RunWithOutput(args ...string) string {
		rOut, out, err := os.Pipe()
	utils.ExpectNoError(err)

	// combine out and errOut
	r.run(out, out, args...)
	err = out.Close()
	utils.ExpectNoError(err)
	outBytes, err := ioutil.ReadAll(rOut)
	utils.ExpectNoError(err)
	err = rOut.Close()
	utils.ExpectNoError(err)
	answer := string(outBytes)
	coverageOutput := coverageOutputRegex.FindStringSubmatch(answer)
	if len(coverageOutput) == 3 {
		utils.LogInfof("when running %s %s coverage was %s", jx, strings.Join(args, " "), coverageOutput[2])
	}
	answer = coverageOutputRegex.ReplaceAllString(answer, "")
	return strings.TrimSpace(answer)
}
