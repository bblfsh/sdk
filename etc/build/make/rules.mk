# docker runtime commands
DOCKER_CMD ?= docker
DOCKER_BUILD ?= $(DOCKER_CMD) build
DOCKER_RUN ?= $(DOCKER_CMD) run --rm
DOCKER_PUSH ?= $(DOCKER_CMD) push
DOCKER_FILE ?= Dockerfile
DOCKER_FILE_BUILD ?= Dockerfile.build
BUILD_VOLUME_TARGET ?= /opt/driver/src/
BUILD_VOLUME_PATH ?= $(shell pwd)

# build enviroment variables
BUILD_USER ?= bblfsh
BUILD_UID ?= $(shell id -u $(USER))
BUILD_ARGS ?= \
	--build-arg BUILD_USER=$(BUILD_USER) \
	--build-arg BUILD_UID=$(BUILD_UID) \
	--build-arg RUNTIME_NATIVE_VERSION=$(RUNTIME_NATIVE_VERSION)
BUILD_NATIVE_CMD ?= $(DOCKER_RUN) \
	-u $(BUILD_USER):$(BUILD_UID) \
	-v $(BUILD_VOLUME_PATH):$(BUILD_VOLUME_TARGET) \
	$(DOCKER_BUILD_IMAGE)

# if VERBOSE is unset docker build is executed in quite mode
ifeq ($(origin VERBOSE), undefined)
	BUILD_ARGS += -q
endif

# term colors helpers
ccred= \e[0;31m
ccyellow= \e[0;33m
ccreset= \e[0m

all: build

$(DOCKER_BUILD_IMAGE):
	$(DOCKER_BUILD) -f $(DOCKER_FILE_BUILD) $(BUILD_ARGS) -t $@  .

test: $(DOCKER_BUILD_IMAGE)
	$(BUILD_NATIVE_CMD) make test-native

build: $(DOCKER_BUILD_IMAGE)
	$(BUILD_NATIVE_CMD) make build-native; \
	$(DOCKER_BUILD) -f $(DOCKER_FILE) $(BUILD_ARGS) -t $(DOCKER_IMAGE):$(DRIVER_VERSION) .

push: build
	@if [ $(pushdisabled) ]; then \
		echo -e "$(ccred)$(pushdisabled)$(ccreset)"; \
		exit 1; \
	fi

	$(DOCKER_PUSH) $(DOCKER_IMAGE):$(DRIVER_VERSION)


