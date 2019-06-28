package utils

import (
	"fmt"
	"os"
	"strings"
)

// MandatoryEnvVar returns an error if the environment variable is missing
func MandatoryEnvVar(name string) (string, error) {
	answer := os.Getenv(name)
	if len(answer) == 0 {
		return "", fmt.Errorf("Missing environment variable value $%s", name)
	}
	return answer, nil
}

// ReplaceEnvVars replaces all environment variable expressions in the given string
func ReplaceEnvVars(expression string) string {
	environ := os.Environ()
	for _, line := range environ {
		s := strings.SplitN(line, "=", 2)
		if len(s) == 2 {
			name := s[0]
			value := s[1]
			expression = strings.Replace(expression, "$"+name, value, -1)
			expression = strings.Replace(expression, "${"+name+"}", value, -1)
		}
	}
	return expression
}
