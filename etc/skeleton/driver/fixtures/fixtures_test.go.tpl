package fixtures

import (
	"path/filepath"
	"testing"

	"github.com/bblfsh/{{.Manifest.Language}}-driver/driver/normalizer"
	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/fixtures"
	"github.com/bblfsh/sdk/v3/driver/native"
	"github.com/bblfsh/sdk/v3/uast/transformer/positioner"
)

const projectRoot = "../../"

var Suite = &fixtures.Suite{
	Lang: "{{.Manifest.Language}}",
	Ext:  ".ext", // TODO: specify correct file extension for source files in ./fixtures
	Path: filepath.Join(projectRoot, fixtures.Dir),
	NewDriver: func() driver.Native {
		return native.NewDriverAt(filepath.Join(projectRoot, "build/bin/native"), native.UTF8)
	},
	Transforms: normalizer.Transforms,
	//BenchName: "fixture-name", // TODO: specify a largest file
	Semantic: fixtures.SemanticConfig{
		BlacklistTypes: []string{
			// TODO: list native types that should be converted to semantic UAST
		},
	},
	VerifyTokens: []positioner.VerifyToken{
	    // TODO: list nodes that needs to be checked for token correctness
	},
}

func Test{{expName .Manifest.Language}}Driver(t *testing.T) {
	Suite.RunTests(t)
}

func Benchmark{{expName .Manifest.Language}}Driver(b *testing.B) {
	Suite.RunBenchmarks(b)
}
