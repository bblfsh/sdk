DRIVER_DEV_PREFIX := dev
DRIVER_VERSION ?= $(DRIVER_DEV_PREFIX)-$(shell git rev-parse HEAD | cut -c1-7)
GO_RUNTIME_VERSION = $(go version)

DOCKER_IMAGE ?= bblfsh/$(LANGUAGE)-driver
DOCKER_BUILD_IMAGE ?= $(DOCKER_IMAGE)-build

# defined behaviour for builds inside travis-ci
ifneq ($(origin CI), undefined)
    # if we are inside CI, verbose is enabled by default
	VERBOSE := true
endif

# if TRAVIS_TAG defined DRIVER_VERSION is overrided
ifneq ($(TRAVIS_TAG), )
    DRIVER_VERSION := $(TRAVIS_TAG)
endif

# if we are not in tag, the push is disabled
ifeq ($(firstword $(subst -, ,$(DRIVER_VERSION))), $(DRIVER_DEV_PREFIX))
	pushdisabled = "push disabled for development versions"
endif