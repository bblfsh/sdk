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
	Lang: "{{.Manifest.Language}}",
	Ext:  ".ext", // TODO: specify correct file extension for source files in ./fixtures
	Path: filepath.Join(projectRoot, fixtures.Dir),
	NewDriver: func() driver.BaseDriver {
		return driver.NewExecDriverAt(filepath.Join(projectRoot, "build/bin/native"))
	},
	Transforms: driver.Transforms{
		Preprocess: normalizer.Preprocess,
		Normalize:  normalizer.Normalize,
		Native:     normalizer.Native,
		Code:       normalizer.Code,
	},
	//BenchName: "fixture-name", // TODO: specify a largest file
	Semantic: fixtures.SemanticConfig{
		BlacklistTypes: []string{
			// TODO: list native types that should be converted to semantic UAST
		},
	},
}

func Test{{expName .Manifest.Language}}Driver(t *testing.T) {
	Suite.RunTests(t)
}

func Benchmark{{expName .Manifest.Language}}Driver(b *testing.B) {
	Suite.RunBenchmarks(b)
}
