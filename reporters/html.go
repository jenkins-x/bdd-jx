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
	readLine := strings.TrimSuffix(specSummary.ComponentTexts[1], "\n")
	if !specSummary.Skipped() {
		r.SpecFailures[readLine] = append(r.SpecFailures[readLine], specSummary.Failed())
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
	data := []SpecResult{}
	for _, i := range r.SpecNames {
		result := SpecResult{Name: i}
		if Contains(r.SpecFailures[i], true) {
			result.Fail = true
		} else {
			result.Fail = false
		}
		data = append(data, result)
	}
	tmpl := template.Must(template.ParseFiles("./templates/layout.html"))
	ti := time.Now()
	time := ti.Format("2006_01_02_15_04_05")
	fileName := RemoveSpaces(ToSnakeCase("report_" + time))

	f, err := os.Create("./reports/" + fileName + ".html")
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
