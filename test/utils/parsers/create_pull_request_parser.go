package parsers

import (
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

var createPullRequestOutputRegex = regexp.MustCompile(`^https:\/\/.*\/pull.*\/([0-9]*)$`)

type CreatePullRequest struct {
	Provider          string
	Owner             string
	Repository        string
	PullRequestNumber int
	Url               string
}

func ParseJxCreatePullRequest(s string) (*CreatePullRequest, error) {
	s = strings.TrimPrefix(s, "Created Pull Request: ")
	parts := createPullRequestOutputRegex.FindStringSubmatch(s)
	if len(parts) != 5 {
		return nil, errors.Errorf("Unable to parse %s as output from jx create pull request", s)
	}
	prn, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, errors.Wrapf(err, "converting pull request number %s to int, entire output was %s", parts[4], s)
	}
	return &CreatePullRequest{
		Provider:          parts[1],
		Owner:             parts[2],
		Repository:        parts[3],
		PullRequestNumber: prn,
		Url:               s,
	}, nil
}
