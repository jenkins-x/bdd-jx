package reporters

import (
	"bufio"
	"html/template"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

// ReporterHTML struct for containing all spec results during suite run
type ReporterHTML struct {
	SpecFailures map[string][]bool `json:"spec_failures"`
	SpecNames    []string          `json:"spec_names"`
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

// SpecSuiteWillBegin Implements ginkgo.Reporter interface. Does not run in parallel mode.
func (r *ReporterHTML) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

// BeforeSuiteDidRun Implements ginkgo.Reporter interface
func (r *ReporterHTML) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {}

// SpecWillRun Implements ginkgo.Reporter interface
func (r *ReporterHTML) SpecWillRun(specSummary *types.SpecSummary) {}

// SpecDidComplete Implements ginkgo.Reporter interface
func (r *ReporterHTML) SpecDidComplete(specSummary *types.SpecSummary) {
	if !specSummary.Skipped() {
		if !specSummary.Passed() {
			readLine := strings.TrimSuffix(specSummary.ComponentTexts[1], "\n")
			r.SpecFailures[readLine] = append(r.SpecFailures[readLine], true)
		}
	}
}

// AfterSuiteDidRun Implements ginkgo.Reporter interface.
func (r *ReporterHTML) AfterSuiteDidRun(setupSummary *types.SetupSummary) {}

// SpecSuiteDidEnd Implements ginkgo.Reporter interface. Does not run in parallel mode.
func (r *ReporterHTML) SpecSuiteDidEnd(summary *types.SuiteSummary) {}

// CreateHTMLReport Creates a HTML report from a final report assembled in the SynchronizedAfterSuite hook.
func (r *ReporterHTML) CreateHTMLReport() {
	r.SpecNames = []string{}
	for k := range r.SpecFailures {
		r.SpecNames = append(r.SpecNames, k)
	}
	sort.Strings(r.SpecNames)
	data := SpecResults{}
	for _, i := range r.SpecNames {
		result := SpecResult{Name: i}
		if Contains(r.SpecFailures[i], true) {
			result.Fail = true
		} else {
			result.Fail = false
		}
		data.Results = append(data.Results, result)
	}
	ti := time.Now().UTC()
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
}
