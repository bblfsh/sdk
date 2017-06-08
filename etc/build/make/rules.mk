BUILD_ID := $(shell date +"%m-%d-%Y_%H_%M_%S")
BUILD_PATH := $(location)/build

RUN := $(sdklocation)/etc/run.sh
RUN_VERBOSE := VERBOSE=1 $(RUN)
RUN_IT := $(sdklocation)/etc/it.bash

# docker runtime commands
DOCKER_CMD ?= docker
DOCKER_BUILD ?= $(DOCKER_CMD) build
DOCKER_RUN ?= $(DOCKER_CMD) run --rm -t
DOCKER_TAG ?= $(DOCKER_CMD) tag
DOCKER_PUSH ?= $(DOCKER_CMD) push

BUILD_VOLUME_TARGET ?= /opt/driver/src/
BUILD_VOLUME_PATH ?= $(location)

DOCKER_FILE_$(DOCKER_IMAGE_VERSIONED) ?= $(location)/Dockerfile.tpl
DOCKER_FILE_$(DOCKER_BUILD_DRIVER_IMAGE) ?= $(sdklocation)/etc/build/Dockerfile.build.$(RUNTIME_OS).tpl
DOCKER_FILE_$(DOCKER_BUILD_NATIVE_IMAGE) ?= $(location)/Dockerfile.build.tpl

# list of images to build
BUILD_IMAGE=$(DOCKER_BUILD_NATIVE_IMAGE) $(DOCKER_BUILD_DRIVER_IMAGE) $(DOCKER_IMAGE_VERSIONED)

# golang runtime commands
GO_CMD = go
GO_TEST = $(GO_CMD) test -v
GO_LDFLAGS = -X main.version=$(DRIVER_VERSION) -X main.build=$(BUILD_ID)

# build enviroment variables
BUILD_USER ?= bblfsh
BUILD_UID ?= $(shell id -u $(USER))
BUILD_ARGS ?=
BUILD_NATIVE_CMD ?= $(DOCKER_RUN) \
	-u $(BUILD_USER):$(BUILD_UID) \
	-v $(BUILD_VOLUME_PATH):$(BUILD_VOLUME_TARGET) \
	-e ENVIRONMENT=$(DOCKER_BUILD_NATIVE_IMAGE) \
	-e HOST_PLATFORM=$(shell uname) \
	$(DOCKER_BUILD_NATIVE_IMAGE)

BUILD_DRIVER_CMD ?= $(DOCKER_RUN) \
	-u $(BUILD_USER):$(BUILD_UID) \
	-v $(BUILD_VOLUME_PATH):$(BUILD_VOLUME_TARGET) \
	-v $(GOPATH):/go \
	-e ENVIRONMENT=$(DOCKER_BUILD_DRIVER_IMAGE) \
	-e HOST_PLATFORM=$(shell uname) \
	$(DOCKER_BUILD_DRIVER_IMAGE)

# if VERBOSE is unset docker build is executed in quite mode
ifeq ($(origin VERBOSE), undefined)
	BUILD_ARGS += -q
endif

ALLOWED_IN_DOCKERFILE = \
	LANGUAGE \
	RUNTIME_NATIVE_VERSION RUNTIME_GO_VERSION \
	BUILD_USER BUILD_UID BUILD_PATH \
	DOCKER_IMAGE DOCKER_IMAGE_VERSIONED DOCKER_BUILD_NATIVE_IMAGE

# we export the variable to allow envsubst, substitute the vars in the
# Dockerfiles
export

all: build

check-gopath:
ifeq ($(GOPATH),)
	$(error GOPATH is not defined)
endif

$(BUILD_PATH):
	@$(RUN) mkdir -p $(BUILD_PATH)

$(BUILD_IMAGE):
	@$(eval TEMP_FILE := $(shell mktemp -p $(tmplocation)))
	@eval "envsubst '$(foreach v,$(ALLOWED_IN_DOCKERFILE),\$${$(v)})' < $(DOCKER_FILE_$@) > $(TEMP_FILE)"
	@$(RUN) $(DOCKER_BUILD) $(BUILD_ARGS) -t $(call unescape_docker_tag,$@) -f $(TEMP_FILE) .
	@rm $(TEMP_FILE)

test: | validate test-native test-driver
test-native: $(DOCKER_BUILD_NATIVE_IMAGE)
	@$(RUN_VERBOSE) $(BUILD_NATIVE_CMD) make test-native-internal

test-driver: | check-gopath $(BUILD_PATH) $(DOCKER_BUILD_NATIVE_IMAGE) $(DOCKER_BUILD_DRIVER_IMAGE)
	@$(RUN_VERBOSE) $(BUILD_DRIVER_CMD) make test-driver-internal

test-driver-internal: build-native-internal
	@cd driver; \
	$(RUN_VERBOSE) $(GO_TEST) ./...

build: | build-native build-driver $(DOCKER_IMAGE_VERSIONED)
build-native: $(BUILD_PATH) $(DOCKER_BUILD_NATIVE_IMAGE)
	@$(RUN) $(BUILD_NATIVE_CMD) make build-native-internal

build-driver: | check-gopath $(BUILD_PATH) $(DOCKER_BUILD_DRIVER_IMAGE) build-native
	@$(RUN) $(BUILD_DRIVER_CMD) make build-driver-internal

build-driver-internal:
	@cd driver; \
	LDFLAGS="$(GO_LDFLAGS)" $(RUN) $(GO_CMD) build -o $(BUILD_PATH)/driver .; \

integration-test: build
	@$(RUN_VERBOSE) $(RUN_IT) "$(call unescape_docker_tag,$(DOCKER_IMAGE_VERSIONED))"

push: build
	$(if $(pushdisabled),$(error $(pushdisabled)))

	@if [ "$$DOCKER_USERNAME" != "" ]; then \
		$(DOCKER_CMD) login -u="$$DOCKER_USERNAME" -p="$$DOCKER_PASSWORD"; \
	fi;

	@$(RUN) $(DOCKER_TAG) $(call unescape_docker_tag,$(DOCKER_IMAGE_VERSIONED)) \
		$(call unescape_docker_tag,$(DOCKER_IMAGE)):latest
	@$(RUN) $(DOCKER_PUSH) $(call unescape_docker_tag,$(DOCKER_IMAGE_VERSIONED))
	@$(RUN) $(DOCKER_PUSH) $(call unescape_docker_tag,$(DOCKER_IMAGE):latest)

docgen:
	- echo "TODO: generate markdown table of the rules and store it at ANNOTATIONS.md"

validate:
	@$(RUN) $(bblfsh-sdk) update --dry-run

clean:
	@$(RUN) rm -rf $(BUILD_PATH)
