package apps

import (
	"github.com/jenkins-x/bdd-jx/test/utils"
)

var (
	includeApps = "jx-app-jacoco:0.0.100"

	uiAppName    = "jx-app-ui"
	uiAppVersion = utils.GetEnv("JX_APP_VERSION", "0.0.59")

	// Timeout for waiting for jx add app to complete
	timeoutAppTests = utils.GetTimeoutFromEnv("BDD_TIMEOUT_APP_TESTS", 60)
)
