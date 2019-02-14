package runner

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	jx             = "jx"
	sessionTimeout = 5 * time.Minute
)

// Runner runs a jx command
type JxRunner struct {
	cwd string
}

// New creates a new jx command runnner
func New(cwd string) *JxRunner {
	return &JxRunner{
		cwd: cwd,
	}
}

// Run runs a jx command
func (r *JxRunner) Run(args ...string) {
	command := exec.Command(jx, args...)
	command.Dir = r.cwd
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	session.Wait(sessionTimeout)
	Eventually(session).Should(gexec.Exit(0))
}
