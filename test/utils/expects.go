package utils

import (
	"github.com/jenkins-x/jx/v2/pkg/util"
	. "github.com/onsi/gomega"
)

// ExpectNoError asserts that the error should not not occur
func ExpectNoError(err error) {
	if err != nil {
		LogInfof("FAILED got unexpected error: \n\n%s\n", util.ColorError(err.Error()))
	}
	Expect(err).ShouldNot(HaveOccurred(), "error is printed in log")
}
