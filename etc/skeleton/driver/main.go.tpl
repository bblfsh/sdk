package main

import (
	"fmt"
	"os"

	"github.com/bblfsh/sdk"
	"github.com/bblfsh/sdk/protocol"

	_ "github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"
)

var version string
var build string

func main() {
    protocol.DriverMain(version, build)
}
