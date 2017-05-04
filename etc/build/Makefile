# This is the entrypoint of the bblfsh make SDK, the SDK should be stored at
# `sdklocation`, generated executing `bblfsh-sdk prepare-build`. Many different
# vars as `LANGUAGE` are required, this variables are extracting from the
# `manifest.toml` at the root and the project translating it to a make include
# at `manifest` using `bblfsh-sdk manifest` tool.

location = $(shell pwd)
sdklocation = $(location)/.sdk
manifest = $(sdklocation)/make/manifest.mk
tmplocation = $(sdklocation)/tmp
$(shell mkdir -p $(tmplocation))

bblfsh-sdk := $(shell command -v bblfsh-sdk 2> /dev/null)
bblfsh-tools := $(shell command -v bblfsh-tools 2> /dev/null)
in-container := $(shell echo $$ENVIRONMENT)
host-platform := $(shell echo $$HOST_PLATFORM)

ifdef bblfsh-tools # run only with Golang
    ifdef in-container
    ifneq ($(host-platform),Linux)
        bblfsh-tools :=  go run /go/src/github.com/bblfsh/sdk/cli/bblfsh-tools/main.go
    endif
    endif

    $(shell $(bblfsh-tools) manifest > $(manifest))
endif

include $(sdklocation)/make/manifest.mk
include $(sdklocation)/make/functions.mk
include $(sdklocation)/make/bootstrap.mk
include $(sdklocation)/make/help.mk
include $(sdklocation)/make/rules.mk
sdkloaded = true
