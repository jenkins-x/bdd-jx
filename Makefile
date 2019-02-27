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
GINKGO := GO111MODULE=on ginkgo $(GINKGO_ARGS)
ROOT_PACKAGE := $(shell $(GO) list .)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
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
	$(GINKGO) --slowSpecThreshold=50000

test-parallel: info
	$(GINKGO) --slowSpecThreshold=50000 -p --nodes 8

test-import: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=import

test-app-lifecycle: info
	$(GINKGO) --slowSpecThreshold=50000 --focus="test app" -- -include-apps=jx-app-jacoco:0.0.100

test-app: info
	$(GINKGO) --slowSpecThreshold=50000 --focus="test app" -- -include-apps=$(JX_BDD_INCLUDE_APPS)

test-create-spring: info
	$(GINKGO) --slowSpecThreshold=50000 --focus="create spring"

test-quickstart-android-quickstart: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=android-quickstart

test-quickstart-angular-io-quickstart: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=angular-io-quickstart

test-quickstart-aspnet-app: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=aspnet-app

test-quickstart-golang-http: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=golang-http

test-quickstart-node-http: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=node-http

test-quickstart-open-liberty: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=open-liberty

test-quickstart-python-http: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=python-http

test-quickstart-rails-shopping-cart: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=rails-shopping-cart

test-quickstart-react-quickstart: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=react-quickstart

test-quickstart-rust-http: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=rust-http

test-quickstart-scala-akka-http-quickstart: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=scala-akka-http-quickstart

test-quickstart-spring-boot-http-gradle: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=spring-boot-http-gradle

test-quickstart-spring-boot-rest-prometheus: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=spring-boot-rest-prometheus

test-quickstart-spring-boot-web: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=spring-boot-web

test-quickstart-vertx-rest-prometheus: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=vertx-rest-prometheus

test-upgrade-ingress: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=ingress

test-upgrade-platform: info
	$(GINKGO) --slowSpecThreshold=50000 --focus=platform

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
	$(GO) build $(BUILDFLAGS) -o build/$(NAME) main.go

.PHONY: release clean test
