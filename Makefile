#
# Copyright (C) 2015 Red Hat, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

SHELL := /bin/bash
NAME := bdd-jx
GO := GO111MODULE=on go
GINKGO := GO111MODULE=on ginkgo $(GINKGO_ARGS) -r
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
JX_VERSION := `jx version -n`
SLOW_SPEC_THRESHOLD := 50000

# get list of all available quickstarts, and convert it into a comma delimited list that can be passed into test
JX_BDD_ALL_QUICKSTARTS := $(shell jx get quickstarts --short | sed -e 'H;$${x;s/\n/,/g;s/^,//;p;};d')

PACKAGE_DIRS := $(shell $(GO) list ./test/...)

REV        := $(shell git rev-parse --short HEAD 2> /dev/null  || echo 'unknown')
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
BUILDFLAGS :=
CGO_ENABLED = 0

ifdef DEBUG
BUILDFLAGS += -gcflags "all=-N -l" $(BUILDFLAGS)
endif

VENDOR_DIR=vendor

GITEA_USER ?= testuser
GITEA_PASSWORD ?= testuser
GITEA_EMAIL ?= testuser@acme.com

GIT_PROVIDER_URL ?= https://github.com

GIT_ORGANISATION ?= jenkins-x-tests
GHE_PROVIDER_URL ?= https://github.beescloud.com
GHE_USER ?= dev1
GHE_TOKEN ?= changeme
GHE_EMAIL ?= testuser@acme.com
JX_BDD_INCLUDE_APPS ?= jx-app-jacoco:0.0.100

# Timeouts used in various test steps (in minutes)
# timeout for a given build to complete
BDD_TIMEOUT_BUILD_COMPLETES ?= 20
# the application is deployed into the first automatic staging environment
BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING ?= 10
# given URL returns the given status code within the given time period
BDD_TIMEOUT_URL_RETURNS ?= 5
# timeout for executing jx command line steps
BDD_TIMEOUT_CMD_LINE ?= 1
# timeout for jx add app to complete
BDD_TIMEOUT_APP_TESTS ?= 60
# session wait timeout
BDD_TIMEOUT_SESSION_WAIT ?= 60
# jx runner session timeout
BDD_TIMEOUT_JX_RUNNER_SESSION ?= 5
# devpod timeout
BDD_TIMEOUT_DEVPOD ?= 15

info:
	@echo "JX VERSION INFORMATION"
	@echo
	@echo "$(JX_VERSION)"
	@echo
ifndef GIT_ORGANISATION
	@echo "If you are running locally remember to set GIT_ORGANISATION to your git username in your local environment"
else
	@echo "GIT_ORGANISATION set to $(GIT_ORGANISATION)"
endif
ifdef GIT_PROVIDER_URL
	@echo "GIT_PROVIDER_URL set to ${GIT_PROVIDER_URL}"
endif
ifdef JX_DISABLE_DELETE_APP
	@echo "JX_DISABLE_DELETE_APP is set. Apps created in the test run will NOT be deleted"
else
	@echo "JX_DISABLE_DELETE_APP is not set. If you would like to disable the automatic deletion of apps created by the tests set this variable to TRUE."
endif
ifdef JX_DISABLE_DELETE_REPO
	@echo "JX_DISABLE_DELETE_REPO is set. Repos created in the test run will NOT be deleted"
else
	@echo "JX_DISABLE_DELETE_REPO is not set. If you would like to disable the automatic deletion of repos created by the tests set this variable to TRUE."
endif
ifdef JX_DISABLE_WAIT_FOR_FIRST_RELEASE
	@echo "JX_DISABLE_WAIT_FOR_FIRST_RELEASE is set.  Will not wait for build to be promoted to staging"
else
	@echo "JX_DISABLE_WAIT_FOR_FIRST_RELEASE is not set.  If you would like to disable waiting for the build to be promoted to staging set this variable to TRUE"
endif

ifdef JX_BDD_INCLUDE_APPS
	@echo "JX_BDD_INCLUDE_APPS is set to $(JX_BDD_INCLUDE_APPS)"
else
	@echo "JX_BDD_INCLUDE_APPS is not set."
endif

	@echo "BDD_TIMEOUT_BUILD_COMPLETES timeout value is $(BDD_TIMEOUT_BUILD_COMPLETES)"
	@echo "BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING timeout value is $(BDD_TIMEOUT_BUILD_RUNNING_IN_STAGING)"
	@echo "BDD_TIMEOUT_URL_RETURNS timeout value is $(BDD_TIMEOUT_URL_RETURNS)"
	@echo "BDD_TIMEOUT_CMD_LINE timeout value is $(BDD_TIMEOUT_CMD_LINE)"
	@echo "BDD_TIMEOUT_APP_TESTS timeout value is $(BDD_TIMEOUT_APP_TESTS)"
	@echo "BDD_TIMEOUT_SESSION_WAIT timeout value is $(BDD_TIMEOUT_SESSION_WAIT)"
	@echo "BDD_TIMEOUT_JX_RUNNER timeout value is $(BDD_TIMEOUT_JX_RUNNER)"
	@echo "BDD_TIMEOUT_DEVPOD timeout value is $(BDD_TIMEOUT_DEVPOD)"
ifdef JX_BDD_INCLUDE_APPS
	@echo "JX_BDD_INCLUDE_APPS is set to $(JX_BDD_INCLUDE_APPS)"
else
	@echo "JX_BDD_INCLUDE_APPS is not set."
endif

configure-ghe:
	echo "Setting up GitHub Enterprise support for user $(GHE_USER) email: $(GITEA_EMAIL)"
	jx create git server github $(GHE_PROVIDER_URL) -n GHE
	jx --version
	jx get git server
	#jx delete git server github.com
	jx create git token -n GHE $(GHE_USER) -t $(GHE_TOKEN)


all: test

check: fmt test

test: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-parallel: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -p --nodes 8

test-import: info
	JX_BDD_IMPORTS=node-http,spring-boot-rest-prometheus,spring-boot-http-gradle,golang-http,golang-http-from-jenkins-x-yml $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-app-lifecycle: info
	JX_BDD_INCLUDE_APPS="$(JX_BDD_INCLUDE_APPS)" $(GINKGO) test/apps --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-verify-pods: info
	$(GINKGO) test/step --slowSpecThreshold=50000

test-create-spring: info
	$(GINKGO) test/spring --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-upgrade-ingress: info
	$(GINKGO) test/ingress --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-upgrade-platform: info
	$(GINKGO) test/platform --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-all-quickstarts: info
	JX_BDD_QUICKSTARTS=$(JX_BDD_ALL_QUICKSTARTS) $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

test-supported-quickstarts: info
	JX_BDD_QUICKSTARTS=node-http,spring-boot-http-gradle,golang-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

#targets for individual quickstarts

test-quickstart-dlang-http: info
	JX_BDD_QUICKSTARTS=dlang-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-jenkins-cwp-quickstart: info
	JX_BDD_QUICKSTARTS=jenkins-cwp-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-jenkins-quickstart: info
	JX_BDD_QUICKSTARTS=jenkins-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-node-http-watch-pipeline-activity: info
	JX_BDD_QUICKSTARTS=node-http-watch-pipeline-activity $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-spring-boot-watch-pipeline-activity: info
	JX_BDD_QUICKSTARTS=spring-boot-watch-pipeline-activity $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-android-quickstart: info
	JX_BDD_QUICKSTARTS=android-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-angular-io-quickstart: info
	JX_BDD_QUICKSTARTS=angular-io-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-aspnet-app: info
	JX_BDD_QUICKSTARTS=aspnet-app $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-golang-http: info
	JX_BDD_QUICKSTARTS=golang-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-node-http: info
	JX_BDD_QUICKSTARTS=node-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-open-liberty: info
	JX_BDD_QUICKSTARTS=open-liberty $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-python-http: info
	JX_BDD_QUICKSTARTS=python-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-rails-shopping-cart: info
	JX_BDD_QUICKSTARTS=rails-shopping-cart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-react-quickstart: info
	JX_BDD_QUICKSTARTS=react-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-rust-http: info
	JX_BDD_QUICKSTARTS=rust-http $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-scala-akka-http-quickstart: info
	JX_BDD_QUICKSTARTS=scala-akka-http-quickstart $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-spring-boot-http-gradle: info
	JX_BDD_QUICKSTARTS=spring-boot-http-gradle $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-spring-boot-rest-prometheus: info
	JX_BDD_QUICKSTARTS=spring-boot-rest-prometheus $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-spring-boot-web: info
	JX_BDD_QUICKSTARTS=spring-boot-web $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-quickstart-vertx-rest-prometheus: info
	JX_BDD_QUICKSTARTS=vertx-rest-prometheus $(GINKGO) test/quickstart --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-import-node-http-watch-pipeline-activity: info
	JX_BDD_IMPORTS=node-http-watch-pipeline-activity $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-import-spring-boot-rest-prometheus: info
	JX_BDD_IMPORTS=spring-boot-rest-prometheus $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-import-spring-boot-http-gradle: info
	JX_BDD_IMPORTS=spring-boot-http-gradle $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-import-golang-http: info
	JX_BDD_IMPORTS=golang-http $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-import-golang-http-from-jenkins-x-yml: info
	JX_BDD_IMPORTS=golang-http-from-jenkins-x-yml $(GINKGO) test/_import --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) -focus=batch

test-devpod: info
	$(GINKGO) test/devpods --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD)

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

install: build
	$(GO) get github.com/onsi/ginkgo/ginkgo
	$(GO) get github.com/onsi/gomega/...

bootstrap: install

bootstrap-report:
	npm i -g xunit-viewer

html-report:
	 xunit-viewer --results=reports/junit.xml --output=reports/junit.html --title="BDD Tests"
	 echo "Generated test report at: reports/junit.html"

clean:
	rm -rf build

build:
	$(GO) build $(BUILDFLAGS) ./test/...

.PHONY: release clean test
