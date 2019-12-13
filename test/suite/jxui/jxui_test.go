package jxui_test

import (
	"testing"

	"github.com/jenkins-x/bdd-jx/test/helpers"

	. "github.com/onsi/ginkgo"
)

func TestSuite(t *testing.T) {
	helpers.RunWithReporters(t, "jxui_smoke_tests")
}

var _ = BeforeSuite(helpers.BeforeSuiteCallback)

var _ = SynchronizedAfterSuite(func() {}, helpers.SynchronizedAfterSuiteCallback)
