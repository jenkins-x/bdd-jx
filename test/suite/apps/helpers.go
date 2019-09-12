package apps

import (
	"github.com/jenkins-x/bdd-jx/test/utils"
)

var (
	IncludeApps = "jx-app-jacoco:0.0.100"
	// Timeout for waiting for jx add app to complete
	TimeoutAppTests = utils.GetTimeoutFromEnv("BDD_TIMEOUT_APP_TESTS", 60)
)
