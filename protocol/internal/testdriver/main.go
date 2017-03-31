package main

import (
	"github.com/bblfsh/sdk/protocol"

	"github.com/bblfsh/sdk/etc/skeleton/driver/normalizer" //REPLACE:"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"
)

var version string
var build string

func main() {
	d := protocol.Driver{
		Version:  version,
		Build:    build,
		ToNoder:  normalizer.NativeToNoder,
		Annotate: normalizer.AnnotationRules,
	}
	d.Exec()
}
