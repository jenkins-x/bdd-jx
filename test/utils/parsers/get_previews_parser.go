package parsers

import (
	"github.com/pkg/errors"
	"strings"
)

type Preview struct {
	PullRequest string
	Namespace   string
	Url         string
}

func ParseJxGetPreviews(s string) (map[string]Preview, error) {
	answer := make(map[string]Preview, 0)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	// Trim the header row
	headerFound := false
	for _, line := range lines {
		// Ignore any output before the header
		if strings.HasPrefix(line, "PULL REQUEST") {
			headerFound = true
			continue
		}
		if !headerFound {
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, errors.Errorf("must be three fields in %s, entire output was %s", line, s)
		}
		answer[fields[0]] = Preview{
			PullRequest: fields[0],
			Namespace:   fields[1],
			Url:         fields[2],
		}
	}
	return answer, nil
}
