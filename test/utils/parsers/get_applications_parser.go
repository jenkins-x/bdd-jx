package parsers

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Application struct {
	Name        string
	Version     string
	Url         string
	DesiredPods int
	RunningPods int
}

func ParseJxGetApplications(s string) (map[string]Application, error) {
	answer := make(map[string]Application, 0)
	lines := strings.Split(strings.TrimSpace(s), "\n")
	headerFound := false
	for _, line := range lines {
		// Ignore any output before the header
		if strings.HasPrefix(line, "APPLICATION") {
			headerFound = true
			continue
		}
		if !headerFound {
			continue
		}
		line = strings.TrimSpace(line)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, errors.Errorf("must be at least 3 fields in %s, entire output was %s", line, s)
		}
		var desiredPods, runningPods int
		if len(fields) == 4 {
			pods := strings.Split(fields[2], "/")
			if len(pods) != 2 {
				return nil, errors.Errorf("cannot parse %s as 1/1, entire output was %s", pods, s)
			}
			var err error
			desiredPods, err = strconv.Atoi(pods[1])
			if err != nil {
				return nil, errors.Wrapf(err, "cannot convert %v to integer, entire output was %s", desiredPods, s)
			}
			runningPods, err = strconv.Atoi(pods[0])
			if err != nil {
				return nil, errors.Wrapf(err, "cannot convert %v to integer, entire output was %s", runningPods, s)
			}
		}
		answer[fields[0]] = Application{
			Name:        fields[0],
			Version:     fields[1],
			DesiredPods: desiredPods,
			RunningPods: runningPods,
			Url:         fields[len(fields)-1],
		}
	}
	return answer, nil
}
