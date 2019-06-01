package runner

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/jenkins-x/bdd-jx/test/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	jx = "jx"
)

var (
	// jxRunner session timeout
	TimeoutJxRunner     = utils.GetTimeoutFromEnv("BDD_TIMEOUT_JX_RUNNER", 5)
	coverageOutputRegex = regexp.MustCompile(`(?m:(PASS|FAIL)\n\s*coverage: ([\d\.]*%) of statements in [\w\.\/]*)`)
)

// Runner runs a jx command
type JxRunner struct {
	cwd      string
	timeout  time.Duration
	exitCode int
}

// New creates a new jx command runnner
func New(cwd string, timeout *time.Duration, exitCode int) *JxRunner {
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
	err := r.run(GinkgoWriter, GinkgoWriter, args...)
	utils.ExpectNoError(err)
}

func (r *JxRunner) run(out io.Writer, errOut io.Writer, args ...string) error {
	command := exec.Command(jx, args...)
	command.Dir = r.cwd
	session, err := gexec.Start(command, out, errOut)
	if err != nil {
		return errors.WithStack(err)
	}
	session.Wait(r.timeout)
	Eventually(session).Should(gexec.Exit())
	if session.ExitCode() != r.exitCode {
		return errors.Errorf("expected exit code %d but got %d whilst running command %s %s", r.exitCode, session.ExitCode(), jx, strings.Join(args, " "))
	}
	return nil
}

// Run runs a jx command
func (r *JxRunner) RunWithOutput(args ...string) string {
	rOut, out, err := os.Pipe()
	utils.ExpectNoError(err)

	// combine out and errOut
	rErr := r.run(out, out, args...)
	err = out.Close()
	utils.ExpectNoError(err)
	outBytes, err := ioutil.ReadAll(rOut)
	utils.ExpectNoError(err)
	err = rOut.Close()
	utils.ExpectNoError(err)
	answer := string(outBytes)
	if rErr != nil {
		utils.ExpectNoError(errors.Wrapf(err, "output %s", answer))
	}
	coverageOutput := coverageOutputRegex.FindStringSubmatch(answer)
	if len(coverageOutput) == 3 {
		utils.LogInfof("when running %s %s coverage was %s\n", jx, strings.Join(args, " "), coverageOutput[2])
	}
	answer = coverageOutputRegex.ReplaceAllString(answer, "")
	return strings.TrimSpace(answer)
}
