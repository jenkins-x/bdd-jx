package runner_test

import (
	"testing"

	"github.com/jenkins-x/bdd-jx/test/utils/runner"
	"github.com/stretchr/testify/assert"
)

func TestCoverageRegex(t *testing.T) {
	out := runner.RemoveCoverageText(`APPLICATION           STAGING PODS URL
bdd-spring-1562755897 0.0.1   1/1  http://bdd-spring-1562755897.jx-staging.35.205.74.52.nip.io
PASS
coverage: 5.6% of statements in ./...
`)
	assert.Equal(t, `APPLICATION           STAGING PODS URL
bdd-spring-1562755897 0.0.1   1/1  http://bdd-spring-1562755897.jx-staging.35.205.74.52.nip.io
`, out)

}
