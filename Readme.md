# BDD tests for Jenkins X using [ginkgo](https://github.com/onsi/ginkgo)

## Prerequisits

- __golang__ https://golang.org/doc/install#install
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

There are lots that can be set (see `test/helpers/suite.go`), but the important ones are:

* `GIT_ORGANISATION` to override the git organisation - by default the username of the pipeline user in the connected cluster

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
