ASSETS_PATH := etc
ASSETS := $(shell ls $(ASSETS_PATH))
ASSETS_PACKAGE := assets
BINDATA_FILE := bindata.go
BINDATA_CMD := $(GOPATH)/bin/go-bindata
BINDATA_URL := github.com/jteeuwen/go-bindata

# General
WORKDIR = $(PWD)

# Go parameters
GOCMD = go
GOTEST = $(GOCMD) test -v

# Coverage
COVERAGE_REPORT = coverage.txt
COVERAGE_PROFILE = profile.out
COVERAGE_MODE = atomic

.PHONY: all bindata test test-coverage validate-commit

all: bindata

bindata: $(addsuffix /bindata.go,$(addprefix $(ASSETS_PACKAGE)/,$(ASSETS)))

$(BINDATA_CMD):
	go get $(BINDATA_URL)/...


$(ASSETS_PACKAGE)/%/bindata.go: $(ASSETS_PATH)/% $(ASSETS_PATH)/%/* $(ASSETS_PATH)/%/*/* $(ASSETS_PATH)/%/*/*/* $(BINDATA_CMD)
	$(BINDATA_CMD) \
		-modtime 1 \
		-nocompress \
		-pkg $* \
		-prefix $(ASSETS_PATH)/$* \
		-o $@ \
		$(ASSETS_PATH)/$*/...

test:
	cd $(WORKDIR); \
	$(GOTEST) ./...

test-coverage:
	cd $(WORKDIR); \
	echo "" > $(COVERAGE_REPORT); \
	for dir in `$(GOCMD) list ./... | egrep -v '/(vendor|etc)/'`; do \
		$(GOTEST) $$dir -coverprofile=$(COVERAGE_PROFILE) -covermode=$(COVERAGE_MODE); \
		if [ $$? != 0 ]; then \
			exit 2; \
		fi; \
		if [ -f $(COVERAGE_PROFILE) ]; then \
			cat $(COVERAGE_PROFILE) >> $(COVERAGE_REPORT); \
			rm $(COVERAGE_PROFILE); \
		fi; \
	done; \

validate-commit: bindata
	if git status --untracked-files=no --porcelain | grep -qe '..*' ; then \
		 echo >&2 "generated bindata is out of sync"; \
	         exit 2; \
	fi
