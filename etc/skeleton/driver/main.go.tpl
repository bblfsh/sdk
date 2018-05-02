package main

import (
	_ "github.com/bblfsh/{{.Manifest.Language}}-driver/driver/impl"
	"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"

	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
)

func main() {
	driver.Run(driver.Transforms{
		Native: normalizer.Native,
		Code:   normalizer.Code,
	})
}
