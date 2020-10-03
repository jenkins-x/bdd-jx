package parsers_test

import (
	"strings"
	"testing"

	"github.com/jenkins-x/bdd-jx/test/utils/parsers"
	"github.com/stretchr/testify/assert"
)

func TestGetActivitiesParser(t *testing.T) {
	out := `
STEP                                STARTED AGO DURATION STATUS
cb-kubecd/bdd-gh-1601660823/PR-1 #1
  Release                                 1m20s     1m0s Succeeded
  Preview                                   20s           https://github.com/cb-kubecd/bdd-gh-1601660823/pull/1
    Preview Application                     20s           http://bdd-gh-1601660823-jx.35.184.30.41.nip.io`
	activities, err := parsers.ParseJxGetActivities(out)
	assert.NoError(t, err)
	assert.Len(t, activities, 1)

	key := "cb-kubecd/bdd-gh-1601660823/PR-1 #1"
	activity := activities[key]
	assert.NotNil(t, activity, "no activity found for key %s", key)

	t.Logf("has status %s\n", activity.Status)
	assert.True(t, strings.HasPrefix(activity.Status, "Succeeded"), "should have succeeded %s", activity.Status)

	t.Logf("found activity %#v\n", activity)

}
