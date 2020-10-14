NAME := bdd-jx
GO := GO111MODULE=on go

PACKAGE_DIRS = $(shell $(GO) list ./test/...)

ORG := jenkins-x
ORG_REPO := $(ORG)/$(NAME)
ROOT_PACKAGE := github.com/$(ORG_REPO)
REV := $(shell git rev-parse --short HEAD 2> /dev/null || echo 'unknown')
BRANCH     := $(shell git rev-parse --abbrev-ref HEAD 2> /dev/null  || echo 'unknown')
BUILD_DATE := $(shell date +%Y%m%d-%H:%M:%S)
VERSION ?= $(shell cat VERSION)
GO_VERSION := 1.13

BUILDFLAGS :=

BUILD_TIME_CONFIG_FLAGS ?= ""
TEST_BUILDFLAGS :=  -ldflags "$(BUILD_TIME_CONFIG_FLAGS)"

TESTFLAGS ?= -v

TESTFLAGS += -timeout 2h

SUITE ?= test-create-spring

ifdef DEBUG
BUILDFLAGS += -gcflags "all=-N -l" $(BUILDFLAGS)
endif

export JX_DISABLE_TEST_CHATOPS_COMMANDS=true

all: build

check: fmt

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build

build:
	$(GO) build $(BUILDFLAGS) ./test/...

build-all:
	$(GO) test -run=nope -failfast -short ./test/...

.PHONY: clean test build fmt build-all

### LEGACY TARGETS, use go test when running locally ###

install:
	echo "deprecated"

test-import:
	$(GO) test $(TESTFLAGS) ./test/suite/_import

test-app-lifecycle:
	$(GO) test $(TESTFLAGS) ./test/suite/apps

test-verify-pods:
	$(GO) test $(TESTFLAGS) ./test/suite/step

test-saas:
	$(GO) test $(TESTFLAGS) ./test/suite/saas

test-create-spring:
	$(GO) test $(TESTFLAGS) ./test/suite/spring

test-upgrade-ingress:
	$(GO) test $(TESTFLAGS) ./test/suite/ingress

test-upgrade-boot:
	$(GO) test $(TESTFLAGS) ./test/suite/upgrade

test-upgrade-platform:
	$(GO) test $(TESTFLAGS) ./test/suite/platform

test-supported-quickstarts:
	JX_BDD_QUICKSTARTS= $(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus='(node-http|spring-boot-http-gradle|golang-http)'

test-devpod:
	$(GO) test $(TESTFLAGS) ./test/suite/devpods

test-jxui:
	$(GO) test $(TESTFLAGS) ./test/suite/jxui

test-lighthouse:
	$(GO) test $(TESTFLAGS) ./test/suite/lighthouse

#targets for individual quickstarts
test-quickstart-golang-http:
	$(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus=golang-http

test-quickstart-node-http:
	$(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus=node-http

test-quickstart-spring-boot-http-gradle:
	$(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus=spring-boot-http-gradle

#target individual imports
test-import-golang-http-from-jenkins-x-yml:
	$(GO) test $(TESTFLAGS) ./test/suite/_import -ginkgo.focus=golang-http-from-jenkins-x-yml

test-import-node-http:
	$(GO) test $(TESTFLAGS) ./test/suite/_import -ginkgo.focus=node-http

test-import-spring-boot-rest-prometheus:
	$(GO) test $(TESTFLAGS) ./test/suite/_import -ginkgo.focus=spring-boot-rest-prometheus

test-import-spring-boot-http-gradle:
	$(GO) test $(TESTFLAGS) ./test/suite/_import -ginkgo.focus=spring-boot-http-gradle

test-single-import:
	$(GO) test $(TESTFLAGS) ./test/suite/_import -ginkgo.focus=${BDD_TEST_SINGLE_IMPORT}

testbin:
	$(GO) test $(TESTFLAGS) -c github.com/jenkins-x/bdd-jx/test/suite/main -o build/bddjx $(TEST_BUILDFLAGS)
#	$(GO) test $(TESTFLAGS) -c github.com/jenkins-x/bdd-jx/test/suite/quickstart -o build/bddjx $(TEST_BUILDFLAGS)

linux:
	GOOS=linux GOARCH=amd64 $(GO) test $(TESTFLAGS) -c github.com/jenkins-x/bdd-jx/test/suite/main -o build/linux/bddjx $(TEST_BUILDFLAGS)

bdd-init:
	echo "About to run the BDD tests on the current cluster"
	git config --global credential.helper store
	git config --global --add user.name jenkins-x-bot
	git config --global --add user.email jenkins-x@googlegroups.com
	git config -l
	jx step git validate
	jx step git credentials
	ls -al ~
	cat ~/.gitconfig

bdd: bdd-init $(SUITE)

saas: bdd test-saas

.PHONY: goreleaser
goreleaser:
	step-go-releaser --organisation=$(ORG) --revision=$(REV) --branch=$(BRANCH) --build-date=$(BUILD_DATE) --go-version=$(GO_VERSION) --root-package=$(ROOT_PACKAGE) --version=$(VERSION) --timeout 200m

