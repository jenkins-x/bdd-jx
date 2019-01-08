package bdd_jx

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	gr "github.com/onsi/ginkgo/reporters"

	"github.com/jenkins-x/bdd-jx/reporters"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var reporterHTML *reporters.ReporterHTML

func TestBddJx(t *testing.T) {
	specFailures := make(map[string][]bool)
	reps := []Reporter{}
	reporterHTML = &reporters.ReporterHTML{
		SpecFailures: specFailures,
	}
	reps = append(reps, reporterHTML)
	reps = append(reps, gr.NewJUnitReporter("reports/junit.xml"))

	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "BddJx Suite", reps)
}

var _ = BeforeSuite(func() {
	var err error
	WorkDir, err = ioutil.TempDir("", TempDirPrefix)
	Expect(err).NotTo(HaveOccurred())
	err = os.MkdirAll(WorkDir, 0760)
	Expect(err).NotTo(HaveOccurred())
	Expect(WorkDir).To(BeADirectory())
})

var _ = SynchronizedAfterSuite(func() {
	// Write json report to file for each node...
	j, err := json.Marshal(reporterHTML)
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
	finalReport := reporters.ReporterHTML{
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
		r := reporters.ReporterHTML{SpecFailures: s}
		err = json.Unmarshal(content, &r)
		if err != nil {
			log.Fatal(err)
		}
		for k, v := range r.SpecFailures {
			finalReport.SpecFailures[k] = append(finalReport.SpecFailures[k], v...)
		}
		i++
	}
	// Create HTML report
	finalReport.CreateHTMLReport()
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
