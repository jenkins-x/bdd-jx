package runner

import (
	"fmt"
	"github.com/jenkins-x/bdd-jx/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	id := "jx"
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		id = fmt.Sprintf("%s_%s", id, arg)
	}
	args, err := AddCoverageArgsIfNeeded(args, id)
	utils.ExpectNoError(err)
	command := exec.Command(jx, args...)
	command.Dir = r.cwd
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	utils.ExpectNoError(err)
	session.Wait(r.timeout)
	Eventually(session).Should(gexec.Exit(r.exitCode))
}

func AddCoverageArgsIfNeeded(args []string, id string) ([]string, error){
	if os.Getenv("ENABLE_COVERAGE") == strings.ToLower("true") || os.Getenv("ENABLE_COVERAGE") == strings.ToLower("1") || os.Getenv("ENABLE_COVERAGE") == strings.ToLower("on") {
		reportsDir := os.Getenv("REPORTS_DIR")
		if reportsDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, errors.Wrapf(err, "getting current dir")
			}
			reportsDir = filepath.Join(cwd, "build","reports")
		}
		outFile := filepath.Join(reportsDir, fmt.Sprintf("%s.%s.out", id, uuid.New()))
		utils.LogInfof("Enabling coverage, writing coverage to %s\n", outFile)
		err := os.Setenv("COVER_JX_BINARY", "true")
		if err != nil {
			return nil, errors.Wrapf(err, "setting env var COVER_JX_BINARY to true")
		}
		err = os.MkdirAll(reportsDir, 0700)
		if err != nil {
			return nil, errors.Wrapf(err, "creating coverage dir")
		}
		args = append([]string{"-test.coverprofile", outFile}, args...)
	}
	return args, nil
}