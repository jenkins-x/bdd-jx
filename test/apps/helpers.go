package apps

import (
	"os"

	"github.com/jenkins-x/bdd-jx/test/utils"
)

var (
	IncludeApps = os.Getenv("JX_BDD_INCLUDE_APPS")
	// Timeout for waiting for jx add app to complete
	TimeoutAppTests = utils.GetTimeoutFromEnv("BDD_TIMEOUT_APP_TESTS", 60)
)
