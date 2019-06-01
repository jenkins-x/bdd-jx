package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gr "github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunWithReporters(t *testing.T, suiteId string) {
	reportsDir := os.Getenv("REPORTS_DIR")
	if reportsDir == "" {
		reportsDir = filepath.Join("../", "build", "reports")
	}
	err := os.MkdirAll(reportsDir, 0700)
	if err != nil {
		t.Errorf("cannot create %s because %v", reportsDir, err)
	}
	reporters := make([]Reporter, 0)

	reporters = append(reporters, gr.NewJUnitReporter(filepath.Join(reportsDir, fmt.Sprintf("%s.junit.xml", suiteId))))
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, fmt.Sprintf("Jenkins X E2E tests: %s", suiteId), reporters)
}

var BeforeSuiteCallback = func() {
	var err error
	WorkDir, err = ioutil.TempDir("", TempDirPrefix)
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(WorkDir, 0760)
	Expect(err).NotTo(HaveOccurred())
	Expect(WorkDir).To(BeADirectory())
}

var SynchronizedAfterSuiteCallback = func() {
	// Cleanup workdir as usual
	cleanFlag := os.Getenv("JX_DISABLE_CLEAN_DIR")
	if strings.ToLower(cleanFlag) != "true" {
		os.RemoveAll(WorkDir)
		Expect(WorkDir).ToNot(BeADirectory())
	}
}
