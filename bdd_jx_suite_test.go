package bdd_jx

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/jenkins-x/bdd-jx/reporters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBddJx(t *testing.T) {
	specFailures := make(map[string][]bool)
	reporterHTML := &reporters.ReporterHTML{SpecFailures: specFailures}
	reporters := []Reporter{reporterHTML}
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "BddJx Suite", reporters)
}

var _ = BeforeSuite(func() {
	var err error
	WorkDir, err = ioutil.TempDir("", TempDirPrefix)
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(WorkDir, 0760)
	Expect(err).NotTo(HaveOccurred())
	Expect(WorkDir).To(BeADirectory())
})

var _ = AfterSuite(func() {
	//os.RemoveAll(WorkDir)
	//Expect(WorkDir).ToNot(BeADirectory())
})
