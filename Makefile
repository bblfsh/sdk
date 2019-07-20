# Package configuration
PROJECT := bblfsh-sdk

# Environment
BASE_PATH := $(shell pwd)

# Assets configuration
ASSETS_PATH := $(BASE_PATH)/etc
ASSETS := $(shell ls $(ASSETS_PATH))
ASSETS_PACKAGE := assets
BINDATA_FILE := bindata.go
BINDATA_CMD := go run github.com/kevinburke/go-bindata/go-bindata/

# Go parameters
GO_CMD = go
GO_TEST = $(GO_CMD) test -v
GO_GET = $(GO_CMD) get -v

# Coverage
COVERAGE_REPORT = coverage.txt
COVERAGE_PROFILE = profile.out
COVERAGE_MODE = atomic

all: bindata

bindata: $(ASSETS)

$(ASSETS):
	chmod -R go=r $(ASSETS_PATH)/$@; \
	$(BINDATA_CMD) \
		-pkg $@ \
		-modtime 1 \
		-nocompress \
		-prefix $(ASSETS_PATH)/$@ \
		-o $(ASSETS_PACKAGE)/$@/$(BINDATA_FILE) \
		$(ASSETS_PATH)/$@/...
	$(GO_CMD) fmt ./$(ASSETS_PACKAGE)/...

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
	git status --untracked-files=no --porcelain | grep -qe '..*'; \
	if  [ $$? -eq 0 ] ; then \
		git diff|cat; \
		echo >&2 "generated bindata is out of sync"; \
		echo >&2 "Re-run 'make bindata' and commit the updates."; \
		exit 2; \
	fi

.PHONY: bindata test test-coverage validate-commit driver-tpl $(ASSETS)
