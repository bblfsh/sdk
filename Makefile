# Package configuration
PROJECT := bblfsh-sdk
DEPENDENCIES := github.com/jteeuwen/go-bindata

# Environment
BASE_PATH := $(shell pwd)

# Assets configuration
ASSETS_PATH := $(BASE_PATH)/etc
ASSETS := $(shell ls $(ASSETS_PATH))
ASSETS_PACKAGE := assets
BINDATA_FILE := bindata.go
BINDATA_CMD := go-bindata

# Go parameters
GO_CMD = go
GO_TEST = $(GO_CMD) test -v
GO_GET = $(GO_CMD) get -v -u

# Coverage
COVERAGE_REPORT = coverage.txt
COVERAGE_PROFILE = profile.out
COVERAGE_MODE = atomic

all: bindata

bindata: $(ASSETS)

$(DEPENDENCIES):
	$(GO_GET) $@/...

$(ASSETS): $(DEPENDENCIES)
	$(BINDATA_CMD) \
		-modtime 1 \
		-pkg $@ \
		-prefix $(ASSETS_PATH)/$@ \
		-o $(ASSETS_PACKAGE)/$@/$(BINDATA_FILE) \
		$(ASSETS_PATH)/$@/...

test:
	$(GO_TEST) ./...

test-coverage:
	echo "" > $(COVERAGE_REPORT); \
	for dir in `$(GO_CMD) list ./... | egrep -v '/(vendor|etc)/'`; do \
		$(GO_TEST) $$dir -coverprofile=$(COVERAGE_PROFILE) -covermode=$(COVERAGE_MODE); \
		if [ $$? != 0 ]; then \
			exit 2; \
		fi; \
		if [ -f $(COVERAGE_PROFILE) ]; then \
			cat $(COVERAGE_PROFILE) >> $(COVERAGE_REPORT); \
			rm $(COVERAGE_PROFILE); \
		fi; \
	done

validate-commit: bindata
	if git status --untracked-files=no --porcelain | grep -qe '..*' ; then \
		$(error generated bindata is out of sync); \
	fi
