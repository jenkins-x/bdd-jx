NAME := bdd-jx
GO := GO111MODULE=on go

PACKAGE_DIRS = $(shell $(GO) list ./test/...)

BUILDFLAGS :=

ifdef DEBUG
BUILDFLAGS += -gcflags "all=-N -l" $(BUILDFLAGS)
endif


all: build

check: fmt

fmt:
	@FORMATTED=`$(GO) fmt $(PACKAGE_DIRS)`
	@([[ ! -z "$(FORMATTED)" ]] && printf "Fixed unformatted files:\n$(FORMATTED)") || true

clean:
	rm -rf build

build:
	$(GO) build $(BUILDFLAGS) ./test/...

.PHONY: clean test build fmt

### LEGACY TARGETS, use go test when running locally ###

install:
	echo "deprecated"

test-import:
	$(GO) test $(TESTFLAGS) ./test/suite/_import

test-app-lifecycle:
	$(GO) test $(TESTFLAGS) ./test/suite/apps

test-verify-pods:
	$(GO) test $(TESTFLAGS) ./test/suite/step

test-create-spring:
	$(GO) test $(TESTFLAGS) ./test/suite/spring

test-upgrade-ingress:
	$(GO) test $(TESTFLAGS) ./test/suite/ingress

test-upgrade-platform:
	$(GO) test $(TESTFLAGS) ./test/suite/platform

test-supported-quickstarts:
	JX_BDD_QUICKSTARTS= $(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus=(node-http|spring-boot-http-gradle|golang-http)

test-devpod:
	$(GO) test $(TESTFLAGS) ./test/suite/devpods

#targets for individual quickstarts
test-quickstart-golang-http:
	$(GO) test $(TESTFLAGS) ./test/suite/quickstart -ginkgo.focus=golang-http
