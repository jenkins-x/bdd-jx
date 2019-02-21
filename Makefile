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
GO := go
GINKGO := ginkgo $(GINKGO_ARGS)
ROOT_PACKAGE := $(shell $(GO) list .)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
JX_VERSION := `jx version -n`
SLOW_SPEC_THRESHOLD := 50000

# get list of all available quickstarts, and convert it into a comma delimited list that can be passed into test
JX_BDD_ALL_QUICKSTARTS := $(shell jx get quickstarts --short | sed -e 'H;$${x;s/\n/,/g;s/^,//;p;};d')

PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)

REV        := $(shell git rev-parse --short HEAD 2> /dev/null  || echo 'unknown')
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
BUILDFLAGS := -ldflags \
  " -X $(ROOT_PACKAGE)/version.Version=$(VERSION)\
		-X $(ROOT_PACKAGE)/version.Revision='$(REV)'\
		-X $(ROOT_PACKAGE)/version.Branch='$(BRANCH)'\
		-X $(ROOT_PACKAGE)/version.BuildDate='$(BUILD_DATE)'\
		-X $(ROOT_PACKAGE)/version.GoVersion='$(GO_VERSION)'"
CGO_ENABLED = 0

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

info:
	@echo "JX VERISON INFORMATION"
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
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=import

test-app-lifecycle: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus="test app" -- -include-apps=jx-app-jacoco:0.0.100

test-app: info
	$(GINKGO) --slowSpecThreshold=50000 --focus="test app" -- -include-apps=$(JX_BDD_INCLUDE_APPS)

test-create-spring: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus="create spring"

test-upgrade-ingress: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=ingress

test-upgrade-platform: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=platform

test-all-quickstarts: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=$(JX_BDD_ALL_QUICKSTARTS)

#targets for individual quickstarts

test-quickstart-dlang-http: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=dlang-http

test-quickstart-jenkins-cwp-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=jenkins-cwp-quickstart

test-quickstart-jenkins-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=jenkins-quickstart

test-quickstart-node-http-watch-pipeline-activity: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=node-http-watch-pipeline-activity

test-quickstart-spring-boot-watch-pipeline-activity: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=spring-boot-watch-pipeline-activity

test-quickstart-android-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=android-quickstart

test-quickstart-angular-io-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=angular-io-quickstart

test-quickstart-aspnet-app: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=aspnet-app

test-quickstart-golang-http: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=golang-http

test-quickstart-node-http: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=node-http

test-quickstart-open-liberty: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=open-liberty

test-quickstart-python-http: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=python-http

test-quickstart-rails-shopping-cart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=rails-shopping-cart

test-quickstart-react-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=react-quickstart

test-quickstart-rust-http: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=rust-http

test-quickstart-scala-akka-http-quickstart: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=scala-akka-http-quickstart

test-quickstart-spring-boot-http-gradle: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=spring-boot-http-gradle

test-quickstart-spring-boot-rest-prometheus: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=spring-boot-rest-prometheus

test-quickstart-spring-boot-web: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=spring-boot-web

test-quickstart-vertx-rest-prometheus: info
	$(GINKGO) --slowSpecThreshold=$(SLOW_SPEC_THRESHOLD) --focus=batch -- -include-quickstarts=vertx-rest-prometheus

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

install:
	$(GO) get -u github.com/onsi/ginkgo/ginkgo
	$(GO) get -u github.com/onsi/gomega/...

bootstrap: install

bootstrap-report:
	npm i -g xunit-viewer

html-report:
	 xunit-viewer --results=reports/junit.xml --output=reports/junit.html --title="BDD Tests"
	 echo "Generated test report at: reports/junit.html"

clean:
	rm -rf build

.PHONY: release clean test
