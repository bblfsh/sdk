package fixtures

import (
	"path/filepath"
	"testing"

	"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver/fixtures"
)

const projectRoot = "../../"

var Suite = &fixtures.Suite{
	Lang: "lang-name",
	Ext:  ".ext",
	Path: filepath.Join(projectRoot, fixtures.Dir),
	NewDriver: func() driver.BaseDriver {
		return driver.NewExecDriverAt(filepath.Join(projectRoot, "build/bin/native"))
	},
	Transforms: driver.Transforms{
		Native: normalizer.Native,
		Code:   normalizer.Code,
	},
	//BenchName: "fixture-name", // TODO: specify a largest file
}

func TestXxxDriver(t *testing.T) {
	Suite.RunTests(t)
}

func BenchmarkXxxDriver(b *testing.B) {
	Suite.RunBenchmarks(b)
}
