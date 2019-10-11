package parsers

import (
	"github.com/pkg/errors"
	"strings"
)

type GitServer struct {
	Name string
	Kind string
	Url  string
}

func ParseJxGetGitServer(s string) ([]GitServer, error) {
	answer := make([]GitServer, 0)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	// Trim the header row
	headerFound := false
	for _, line := range lines {
		// Ignore any output before the header
		if strings.HasPrefix(line, "Name") {
			headerFound = true
			continue
		}
		if !headerFound {
			continue
		}
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, errors.Errorf("must be three fields in %s, entire output was %s", line, s)
		}
		answer = append(answer, GitServer{
			Name: fields[0],
			Kind: fields[1],
			Url:  fields[2],
		})
	}
	return answer, nil
}
