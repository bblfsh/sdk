package main

import (
	"fmt"
	"os"

	"github.com/bblfsh/sdk"

	"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer/"
)

var version string
var build string

func main() {
	fmt.Printf("version: %s\nbuild: %s\n", version, build)

	_, err := os.Stat(sdk.NativeBin)
	if err == nil {
		fmt.Println("native: ok")
		return
	}

	fmt.Printf("native: %s\n", err)
}