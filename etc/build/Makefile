# This is the entrypoint of the bblfsh make SDK, the SDK should be stored at
# `sdklocation`, generated executing `bblfsh-sdk prepare-build`. Many different
# vars as `LANGUAGE` are required, this variables are extracting from the
# `manifest.toml` at the root and the project translating it to a make include
# at `manifest` using `bblfsh-sdk manifest` tool.

sdklocation = .sdk
manifest = $(sdklocation)/make/manifest.mk
$(shell bblfsh-sdk manifest --print-env 2>/dev/null > $(manifest))
include $(sdklocation)/make/manifest.mk
include $(sdklocation)/make/bootstrap.mk
include $(sdklocation)/make/rules.mk