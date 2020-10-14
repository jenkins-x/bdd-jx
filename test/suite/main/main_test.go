package main_test

import (
	"os"
	"testing"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	. "github.com/onsi/ginkgo"

	// lets import the tests
	_ "github.com/jenkins-x/bdd-jx/test/suite/quickstart"
	// lets import the tests
	_ "github.com/jenkins-x/bdd-jx/test/suite/spring"
)

// TestSuite runs one or more tests using environment variables to define the tests to run
func TestSuite(t *testing.T) {
	suiteId := os.Getenv("JX_BDD_SUITE")
	if suiteId == "" {
		suiteId = "create_quickstarts"
	}
	t.Logf("running test suite %s\n", suiteId)
	helpers.RunWithReporters(t, suiteId)
}

var _ = BeforeSuite(helpers.BeforeSuiteCallback)

var _ = SynchronizedAfterSuite(func() {}, helpers.SynchronizedAfterSuiteCallback)
