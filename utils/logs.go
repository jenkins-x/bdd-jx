package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/onsi/ginkgo"
)

const (
	infoPrefix = "      "
)

// LogInfo info logging
func LogInfo(message string) {
	fmt.Fprintln(ginkgo.GinkgoWriter, infoPrefix+message)
}

// LogInfof info logging
func LogInfof(format string, args ...interface{}) {
	fmt.Fprintf(ginkgo.GinkgoWriter, infoPrefix+fmt.Sprintf(format, args...))
}

// Color avoids the color string if we should disable colors
func Color(colorText string) string {
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "xterm") || term == "xterm" {
		return colorText
	}
	return ""
}
