package parsers_test

import (
	"testing"

	"github.com/jenkins-x/bdd-jx/test/utils/parsers"
	"github.com/stretchr/testify/assert"
)

func TestGetApplicationsParser(t *testing.T) {
	out := `
WARNING: could not find the current user name user: Current not implemented on linux/amd64
APPLICATION           STAGING PODS URL
bdd-spring-1561456570 0.0.1   1/1  http://bdd-spring-1561456570.bdd-ghe-jx-pr-4153-100-staging.35.205.242.160.nip.io`
	applications, err := parsers.ParseJxGetApplications(out)
	assert.NoError(t, err)
	assert.Len(t, applications, 1)
}
