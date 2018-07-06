# BDD tests for Jenkins X using [ginkgo](https://github.com/onsi/ginkgo)

## Prerequisits

- __golang__ https://golang.org/doc/install#install
- a Jenkins X installation

## Setup

    make bootstrap
will install ginkgo gomega and dep

## Running the BDD tests

If you are running the tests locally you probably want to set:

    export GIT_ORGANISATION="my_cool_github_username"
    
Then to run all the tests in parallel:

    make test-parallel

If you want the sequential version (You may be some time):

    make test

Or you can run an individual spec like this:

    make test-quickstart-golang-http
