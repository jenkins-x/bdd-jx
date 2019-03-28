package utils

import (
	. "github.com/onsi/gomega"
)

// ExpectNoError asserts that the error should not not occur
func ExpectNoError(err error) {
	if err != nil {
		LogInfof("FAILED got unexpected error: %s\n", err.Error())
	}
	Expect(err).ShouldNot(HaveOccurred())
}
