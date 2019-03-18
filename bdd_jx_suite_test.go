package bdd_jx

import (
	"encoding/json"
	"fmt"
	"github.com/jenkins-x/jx/pkg/util"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gr "github.com/onsi/ginkgo/reporters"

	"github.com/jenkins-x/bdd-jx/reporters"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var ReporterTestGrid *reporters.ReporterTestGrid
func init() {
	fmt.Println("initfunc")
}
func TestBddJx(t *testing.T) {
	fmt.Println("TestBddJxStart")

	specFailures := make(map[string][]bool)
	reps := []Reporter{}
	ReporterTestGrid = &reporters.ReporterTestGrid{
		SpecFailures: specFailures,
		OutputDir:    "reports",
	}
	reps = append(reps, ReporterTestGrid)

	artifactsDir := "reports/artifacts"
	os.MkdirAll(artifactsDir, util.DefaultWritePermissions)
	reps = append(reps, gr.NewJUnitReporter(filepath.Join(artifactsDir, "junit_00.xml")))

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "BddJx Suite", reps)
}

var _ = BeforeSuite(func() {
	var err error
	fmt.Println("BeforeSuiteStuff")
	WorkDir, err = ioutil.TempDir("", TempDirPrefix)
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(WorkDir, 0760)
	Expect(err).NotTo(HaveOccurred())
	Expect(WorkDir).To(BeADirectory())
})

var _ = SynchronizedAfterSuite(func() {
	// Write json report to file for each node...
	j, err := json.Marshal(ReporterTestGrid)
	Expect(err).NotTo(HaveOccurred())
	fileName := "reports/report-data-" + strconv.Itoa(config.GinkgoConfig.ParallelNode) + ".json"
	err = ioutil.WriteFile(fileName, j, 0644)
	Expect(err).NotTo(HaveOccurred())
}, func() {
	// Runs on node 1 when all other nodes have completed.
	var err error
	var content []byte
	i := 1
	fileNames := []string{}
	// Read n json report files & build final report from all node reports
	specFailures := make(map[string][]bool)
	finalReport := reporters.ReporterTestGrid{
		SpecFailures: specFailures,
	}
	for {
		fileName := "reports/report-data-" + strconv.Itoa(i) + ".json"
		fileNames = append(fileNames, fileName)
		content, err = ioutil.ReadFile(fileName)
		if err != nil {
			break
		}
		s := make(map[string][]bool)
		r := reporters.ReporterTestGrid{SpecFailures: s}
		err = json.Unmarshal(content, &r)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range r.SpecFailures {
			finalReport.SpecFailures[k] = append(finalReport.SpecFailures[k], v...)
		}
		i++
	}
	finalReport.CreateTestGridReport()
	// Cleanup node report json files
	for _, f := range fileNames {
		os.Remove(f)
	}
	// Cleanup workdir as usual
	cleanFlag := os.Getenv("JX_DISABLE_CLEAN_DIR")
	if strings.ToLower(cleanFlag) != "true" {
		os.RemoveAll(WorkDir)
		Expect(WorkDir).ToNot(BeADirectory())
	}
})
