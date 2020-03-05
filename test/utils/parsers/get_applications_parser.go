package parsers

import (
	"net/url"
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
		// lets ignore any warnings
		if strings.Contains(line, "WARNING") {
			continue
		}
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
			return nil, errors.Errorf("must be at least %d fields in %s, entire output was %s", 3, line, s)
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
		app := Application{
			Name:        fields[0],
			Version:     fields[1],
			DesiredPods: desiredPods,
			RunningPods: runningPods,
		}
		urlString := fields[len(fields)-1]
		if urlString != "" {
			// The last field can end up as "1/1" (or "0/1" etc) for a brief time before the ingress is created. We want
			// to try again when that happens. The easiest way to do so is to parse it and make sure it has a non-empty
			// scheme.
			u, err := url.Parse(urlString)
			if err != nil {
				return nil, errors.Wrapf(err, "parsing URL %s from full output %s", urlString, s)
			}
			if u != nil && u.Scheme != "" {
				app.Url = urlString
			}
		}
		answer[fields[0]] = app
	}
	return answer, nil
}
