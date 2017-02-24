# Manifest contains metadata about the driver. To learn about the different
# values in this manifest refer to:
#   https://github.com/bblfsh/sdk/blob/master/manifest/manifest.go

language = "{{.Language}}"

# status describes the current development status of the driver, the valid
# values for status are: `planning`, `pre-alpha`, `alpha`, `beta`, `stable`,
# `mature` or `inactive`
status = "planning"

# documentation block is use to render the README.md file
[documentation]
description  = """
{{.Language}} driver for [babelfish](https://github.com/bblfsh/server).
"""

[runtime]
# native_version describes the version or versions being use to build and
# execute the native code, you should define at least one. (eg.: "1.8")
native_version = []