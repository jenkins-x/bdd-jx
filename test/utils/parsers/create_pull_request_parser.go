package parsers

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	// CreatedPRLogLinePrefix is the prefix we'll find on a log line that will be parsable
	CreatedPRLogLinePrefix = "Created Pull Request: "
)

var createPullRequestOutputRegex = regexp.MustCompile(`^https:\/\/([^\/]*)\/(?:projects\/)?([^\/]*)\/(?:repos\/)?([^\/]*)\/(?:-\/)?(?:pull|pull-requests|merge_requests)\/([0-9]*)$`)

type CreatePullRequest struct {
	Provider          string
	Owner             string
	Repository        string
	PullRequestNumber int
	Url               string
}

func ParseJxCreatePullRequest(s string) (*CreatePullRequest, error) {
	s = strings.TrimPrefix(s, CreatedPRLogLinePrefix)
	parts := createPullRequestOutputRegex.FindStringSubmatch(s)
	if len(parts) != 5 {
		return nil, errors.Errorf("Unable to parse %s as output from jx create pull request and has parts %#v", s, parts)
	}
	prn, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, errors.Wrapf(err, "converting pull request number %s to int, entire output was %s and has parts %#v", parts[4], s, parts)
	}
	owner := parts[2]
	provider := parts[1]

	// bitbucket server URLs use upper case ProjectKeys instead of the owner name
	if strings.Contains(provider, "bitbucket") {
		owner = strings.ToLower(owner)
	}
	return &CreatePullRequest{
		Provider:          provider,
		Owner:             owner,
		Repository:        parts[3],
		PullRequestNumber: prn,
		Url:               s,
	}, nil
}

func ParseJxCreatePullRequestFromFullLog(s string) (*CreatePullRequest, error) {
	for _, line := range strings.Split(strings.Replace(s, "\r\n", "\n", -1), "\n") {
		if strings.HasPrefix(line, CreatedPRLogLinePrefix) {
			return ParseJxCreatePullRequest(line)
		}
	}
	return nil, fmt.Errorf("could not find %s in log output", CreatedPRLogLinePrefix)
}
