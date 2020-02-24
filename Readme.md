# BDD tests for Jenkins X using [ginkgo](https://github.com/onsi/ginkgo)

## Prerequisits

- __[golang](https://golang.org/doc/install#install)__
- a Jenkins X installation
    
## Running the BDD tests

Then to run all the test suites

    go test -timeout 2h ./test/suite/...
    
Note that as some of the tests take quite a long time it's important to override the default go test timeout of `10m` 

Or you can run a specific suite, for example

    go test -timeout 1h ./test/suite/spring

To enable verbose logging, add `-v`, for example

    go test -timeout 1h -v ./test/suite/spring 


## Environment variables

There are lots that can be set (see `test/helpers/suite.go`).
Most of the settings are quite specific and need to be used explicitly in your test to apply.

|Environment variable                |Use |
|------------------------------------|----|
|BDD_JX                              | Fully qualified path to `jx` binary to use. If not specified `jx` will use the $PATH to find the binary.   |
|BDD_TIMEOUT_APP_TESTS               | Timeout for Apps related test determining the time to wait for `jx` commands to complete. See _apps.go_ |
|BDD_TIMEOUT_BUILD_COMPLETES         | Timeout waiting for a build to complete, for example a quickstart build. |
|BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING| Timeout waiting for a staging build appearing. |
|BDD_TIMEOUT_CMD_LINE                | Timeout waiting for external command to complete. |
|BDD_TIMEOUT_DEVPOD            	     | Timeout waiting for devpod to appear. |
|BDD_TIMEOUT_SESSION_WAIT            | Timeout waiting for `jx` command to complete. |
|BDD_TIMEOUT_URL_RETURNS             | Timeout waiting for a given URL to become available. |
|GHE_PROVIDER_URL                    | ? |
|GHE_TOKEN                           | ? |
|GHE_USER                            | ? |
|GIT_ORGANISATION                    | GitHub organization used as owner for created repositories. |
|GIT_PROVIDER_URL                    | Git provider URL. |
|JX_BDD_INCLUDE_APPS                 | Comma separated list of apps for which to test the app life cycle. Defaults to _jx-app-jacoco:0.0.100_|
|JX_DISABLE_DELETE_APP               | Whether application created via quickstart test should be deleted. |
|JX_DISABLE_DELETE_REPO              | Whether repositories created via quickstart test should be deleted. |
|JX_DISABLE_WAIT_FOR_FIRST_RELEASE   | ? |
|SLOW_SPEC_THRESHOLD                 | Ginkgo threshold for marking a spec as slow. |

### Running tests locally

When trying to run the tests locally against an existing cluster the following variables are in particular interesting:


* `GIT_ORGANISATION` to override the git organisation - by default the username of the pipeline user in the connected cluster
* `BDD_JX` to override the `jx` binary to use for executing `jx` commands
* `JX_HOME` to override the Jenkins X home directory, default _~/.jx_ 
* `KUBECONTEXT` to point to a given cluster


## Debugging tests in your IDE

### Goland

Find the right `_test.go` file for the suite you want to run, and right-click `Run ..._test.go`, or choose `Debug ..._test.go` to debug it.

If it's a long running test, you may need to add the `-timeout 1h` argument, you can do that by editing the `Run` configuration and adding `-timeout 1h` to the `Go Tool Arguments`.

To enable verbose logging edit the `Run` configuration and adding `-v` to the `Go Tool Arguments`. 

### Visual Studio Code

Microsoft already has [excellent general guidance](https://github.com/Microsoft/vscode-go/wiki/Debugging-Go-code-using-VS-Code) for debugging Go in VS Code. Below is an example 
of how to set this up specifically for these BDD tests.

1. From the drop-down menus, select _Debug_ --> _Open Configurations_. This should create a `launch.json` file for you, if there wasn't one already set up.

2. Some of the default settings for this file will be fine, but others need to be changed e.g. to set the timeut.. VS Code's Intellisense feature will let you hover over these settings and see possibilities. 
