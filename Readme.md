# BDD tests for Jenkins X using [ginkgo](https://github.com/onsi/ginkgo)

## Prerequisits

- __golang__ https://golang.org/doc/install#install
- a Jenkins X installation

## Setup

    make bootstrap

will install ginkgo gomega and dep

To enable the `report` goal you will also need to install [xunit-viewer](https://github.com/lukejpreston/xunit-viewer) via:

    make bootstrap-report
    
## Running the BDD tests

If you are running the tests locally you probably want to set:

    export GIT_ORGANISATION="my_cool_github_username"
    
Then to run all the tests in parallel:

    make test-parallel

If you want the sequential version (You may be some time):

    make test

Or you can run an individual spec like this:

    make test-quickstart-golang-http

To add the HTML report generation add the goal `html-report` like this:

    make test html-report

To enable verbose logging do this before running `make`

    export GINKGO_ARGS=-v

## Environment variables

* `GINKGO_ARGS` to pass in any [ginkgo command line arguments](http://onsi.github.io/ginkgo/#the-ginkgo-cli, like `-v` for verbose logging
* `GIT_PROVIDER_URL` the git provider URL to test against. e.g. your GitHub Enterprise or BitBucket URL
* `JX_DISABLE_CLEAN_DIR` set to `true` to disable cleaning up of the temporary work directories 
* `JX_DISABLE_DELETE_APP` set to `true` to disable deleting of the app from Jenkins X after a test
* `JX_DISABLE_DELETE_REPO` set to `true` to disable deleting of the repo from Jenkins X after a test
* `JX_DISABLE_TEST_PULL_REQUEST` set to `true` to disable testing the PR workflow
* `JX_DISABLE_WAIT_FOR_FIRST_RELEASE` set to `true` to disable waiting for the first release pipeline to complete. Handy if you are testing/debugging the Pull Request flow as it speeds up the test
