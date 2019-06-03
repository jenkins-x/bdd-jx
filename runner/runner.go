package runner

import (
	"github.com/jenkins-x/bdd-jx/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"os/exec"
	"time"
)

const (
	jx = "jx"
)

var (
	// jxRunner session timeout
	TimeoutJxRunner = utils.GetTimeoutFromEnv("BDD_TIMEOUT_JX_RUNNER", 5)
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
		cwd: cwd,
		timeout: *timeout,
		exitCode: exitCode,
	}
}

// Run runs a jx command
func (r *JxRunner) Run(args ...string) {
	command := exec.Command(jx, args...)
	command.Dir = r.cwd
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	utils.ExpectNoError(err)
	session.Wait(r.timeout)
	Eventually(session).Should(gexec.Exit(r.exitCode))
}
