package reporters

import (
	"encoding/json"
	"github.com/jenkins-x/jx/pkg/util"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

const (
	resultSuccess = "SUCCESS"
	resultFailure = "FAILURE"
)

// ReporterTestGrid struct for containing all spec results during suite run
type ReporterTestGrid struct {
	OutputDir    string
	SpecFailures map[string][]bool `json:"spec_failures"`
	SpecNames    []string          `json:"spec_names"`
	StartTime    time.Time
	EndTime      time.Time
}

// SpecResults struct for writing results to template
type SpecResults struct {
	Results       []SpecResult
	TimeCompleted string
}

// SpecResult struct for writing results to template
type SpecResult struct {
	Name string
	Fail bool
}

// TestGridStarted the started.json data
type TestGridStarted struct {
	Timestamp int64 `json:"timestamp"`
}

// TestGridFinished the finished.json data
type TestGridFinished struct {
	Timestamp int64  `json:"timestamp"`
	Passed    bool   `json:"passed"`
	Result    string `json:"result"`
}

// SpecSuiteWillBegin Implements ginkgo.Reporter interface. Does not run in parallel mode.
func (r *ReporterTestGrid) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

// BeforeSuiteDidRun Implements ginkgo.Reporter interface
func (r *ReporterTestGrid) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	r.StartTime = time.Now()

	reportJSON := TestGridStarted{
		Timestamp: r.StartTime.Unix(),
	}
	r.WriteJSONFile("started.json", &reportJSON)
}

// SpecWillRun Implements ginkgo.Reporter interface
func (r *ReporterTestGrid) SpecWillRun(specSummary *types.SpecSummary) {}

// SpecDidComplete Implements ginkgo.Reporter interface
func (r *ReporterTestGrid) SpecDidComplete(specSummary *types.SpecSummary) {
	if !specSummary.Skipped() {
		if !specSummary.Passed() {
			readLine := strings.TrimSuffix(specSummary.ComponentTexts[1], "\n")
			r.SpecFailures[readLine] = append(r.SpecFailures[readLine], true)
		}
	}
}

// AfterSuiteDidRun Implements ginkgo.Reporter interface.
func (r *ReporterTestGrid) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

// SpecSuiteDidEnd Implements ginkgo.Reporter interface. Does not run in parallel mode.
func (r *ReporterTestGrid) SpecSuiteDidEnd(summary *types.SuiteSummary) {
	r.EndTime = time.Now()

	reportJSON := TestGridFinished{
		Timestamp: r.StartTime.Unix(),
		Passed:    true,
		Result:    resultSuccess,
	}

	failed := summary.NumberOfFailedSpecs
	if failed > 0 {
		reportJSON.Passed = false
		reportJSON.Result = resultFailure
	}

	r.WriteJSONFile("finished.json", &reportJSON)

	/*
		r.SpecNames = []string{}
		for k := range r.SpecFailures {
			r.SpecNames = append(r.SpecNames, k)
		}
		sort.Strings(r.SpecNames)
		data := SpecResults{}

		reportJSON.Passed = true
		for _, i := range r.SpecNames {
			result := SpecResult{Name: i}
			if Contains(r.SpecFailures[i], true) {
				result.Fail = true
			} else {
				result.Fail = false
				reportJSON.Passed = false
			}
			data.Results = append(data.Results, result)
		}

	*/

	/*	ti := time.Now().UTC()
		time := ti.Format("2 Jan 2006 15:04 UTC")
		data.TimeCompleted = time
		tmpl := template.Must(template.ParseFiles("./templates/layout.html"))

		f, err := os.Create("./reports/build-status.html")
		if err != nil {
			log.Fatal("Execute: ", err)
			return
		}
		w := bufio.NewWriter(f)
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Fatal("Execute: ", err)
			return
		}
		w.Flush()
		f.Close()
	*/
}

// CreateTestGridReport Creates the TestGrid report files
func (r *ReporterTestGrid) CreateTestGridReport() {
}

func (r *ReporterTestGrid) WriteJSONFile(fileName string, v interface{}) {
	fullName := fileName
	if r.OutputDir != "" {
		fullName = filepath.Join(r.OutputDir, fileName)
	}

	data, err := json.Marshal(v)
	if err != nil {
		log.Fatal("Failed to marshal JSON data: ", v, " for file file: ", fullName, " due to: ", err)
		return
	}

	err = ioutil.WriteFile(fullName, data, util.DefaultWritePermissions)
	if err != nil {
		log.Fatal("Failed to write file: ", fullName, " due to: ", err)
	}
}
