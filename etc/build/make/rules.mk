BUILD_ID := $(shell date +"%m-%d-%Y_%H_%M_%S")
BUILD_PATH := $(location)/build

# docker runtime commands
DOCKER_CMD ?= docker
DOCKER_BUILD ?= $(DOCKER_CMD) build
DOCKER_RUN ?= $(DOCKER_CMD) run --rm -t
DOCKER_TAG ?= $(DOCKER_CMD) tag
DOCKER_PUSH ?= $(DOCKER_CMD) push

BUILD_VOLUME_TARGET ?= /opt/driver/src/
BUILD_VOLUME_PATH ?= $(location)

DOCKER_FILE_$(DOCKER_IMAGE_VERSIONED) ?= $(location)/Dockerfile
DOCKER_FILE_$(DOCKER_BUILD_DRIVER_IMAGE) ?= $(sdklocation)/etc/Dockerfile.build.$(RUNTIME_OS)
DOCKER_FILE_$(DOCKER_BUILD_NATIVE_IMAGE) ?= $(location)/Dockerfile.build

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
	$(DOCKER_BUILD_NATIVE_IMAGE)

BUILD_DRIVER_CMD ?= $(DOCKER_RUN) \
	-u $(BUILD_USER):$(BUILD_UID) \
	-v $(BUILD_VOLUME_PATH):$(BUILD_VOLUME_TARGET) \
	-v $(GOPATH):/go \
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

# term colors helpers
ccred= \e[0;31m
ccyellow= \e[0;33m
ccreset= \e[0m

# we export the variable to allow envsubst, substitute the vars in the
# Dockerfiles
export

all: build

$(BUILD_PATH):
	mkdir -p $(BUILD_PATH)

$(BUILD_IMAGE):
	@$(eval TEMP_FILE := $(shell mktemp -p $(tmplocation)))
	@eval "envsubst '$(foreach v,$(ALLOWED_IN_DOCKERFILE),\$${$(v)})' < $(DOCKER_FILE_$@) > $(TEMP_FILE)"
	$(DOCKER_BUILD) $(BUILD_ARGS) -t $(call unescape_docker_tag,$@) -f $(TEMP_FILE) .

test: | test-native test-driver
test-native: $(DOCKER_BUILD_NATIVE_IMAGE)
	$(BUILD_NATIVE_CMD) make test-native-internal

test-driver: | $(BUILD_PATH) $(DOCKER_BUILD_NATIVE_IMAGE) $(DOCKER_BUILD_DRIVER_IMAGE)
	$(BUILD_DRIVER_CMD) make test-driver-internal

test-driver-internal: build-native-internal
	cd driver; \
	$(GO_TEST) ./...

build: | build-native build-driver $(DOCKER_IMAGE_VERSIONED)
build-native: $(BUILD_PATH) $(DOCKER_BUILD_NATIVE_IMAGE)
	$(BUILD_NATIVE_CMD) make build-native-internal

build-driver: | $(BUILD_PATH) $(DOCKER_BUILD_DRIVER_IMAGE) build-native
	$(BUILD_DRIVER_CMD) make build-driver-internal

build-driver-internal:
	cd driver; \
	$(GO_CMD) build --ldflags '$(GO_LDFLAGS)' -o $(BUILD_PATH)/driver .; \

push: build
	@if [ $(pushdisabled) ]; then \
		echo -e "$(ccred)$(pushdisabled)$(ccreset)"; \
		exit 1; \
	fi

	$(DOCKER_TAG) $(call unescape_docker_tag,$(DOCKER_IMAGE_VERSIONED)) \
		$(call unescape_docker_tag,$(DOCKER_IMAGE)):latest
	$(DOCKER_PUSH) $(DOCKER_IMAGE_VERSIONED)
	$(DOCKER_PUSH) $(DOCKER_IMAGE):latest

clean:
	rm -rf $(BUILD_PATH)