package driver

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v1/manifest"
	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"

	"github.com/stretchr/testify/require"
)

func init() {
	ManifestLocation = "internal/native/manifest.toml"
}

func TestDriverParserParse(t *testing.T) {
	require := require.New(t)
	NativeBinary = "internal/native/mock"

	d, err := NewDriver(&uast.ObjectToNode{}, nil)
	require.NoError(err)
	require.NotNil(d)

	err = d.Start()
	require.NoError(err)

	r := d.Parse(&protocol.ParseRequest{
		Language: "fixture",
		Filename: "foo.f",
		Content:  "foo",
	})

	require.NotNil(r)
	require.Equal(len(r.Errors), 0)
	require.Equal(r.Status, protocol.Ok)
	require.Equal(r.Language, "fixture")
	require.Equal(r.Filename, "foo.f")
	require.Equal(r.Elapsed.Nanoseconds() > 0, true)
	require.Equal(r.UAST.String(), " "+
		"{\n"+
		".  Roles: Unannotated\n"+
		".  Properties: {\n"+
		".  .  key: val\n"+
		".  }\n"+
		"}\n",
	)

	err = d.Stop()
	require.NoError(err)
}

func TestDriverParserParse_MissingLanguage(t *testing.T) {
	require := require.New(t)

	d, err := NewDriver(&uast.ObjectToNode{}, nil)
	require.NoError(err)
	require.NotNil(d)

	err = d.Start()
	require.NoError(err)

	r := d.Parse(&protocol.ParseRequest{
		Content: "foo",
	})

	require.NotNil(r)
	require.Equal(len(r.Errors), 1)
	require.Equal(r.Status, protocol.Fatal)
	require.Equal(r.Elapsed.Nanoseconds() > 0, true)
	require.Nil(r.UAST)

	err = d.Stop()
	require.NoError(err)
}
func TestDriverParserParse_Malfunctioning(t *testing.T) {
	require := require.New(t)
	NativeBinary = "echo"

	d, err := NewDriver(&uast.ObjectToNode{}, nil)
	require.NoError(err)
	require.NotNil(d)

	err = d.Start()
	require.NoError(err)

	r := d.Parse(&protocol.ParseRequest{
		Language: "fixture",
		Content:  "foo",
	})

	require.NotNil(r)

	require.Equal(r.Status, protocol.Fatal)
	require.Equal(r.Elapsed.Nanoseconds() > 0, true)
	require.Equal(len(r.Errors), 1)
	require.Nil(r.UAST)

	err = d.Stop()
	require.NoError(err)
}

func TestDriverParserNativeParse(t *testing.T) {
	require := require.New(t)
	NativeBinary = "internal/native/mock"

	d, err := NewDriver(&uast.ObjectToNode{}, nil)
	require.NoError(err)
	require.NotNil(d)

	err = d.Start()
	require.NoError(err)

	r := d.NativeParse(&protocol.NativeParseRequest{
		Language: "fixture",
		Content:  "foo",
	})

	require.NotNil(r)
	require.Equal(len(r.Errors), 0)
	require.Equal(r.Status, protocol.Ok)
	require.Equal(r.Language, "fixture")
	require.Equal(r.Elapsed.Nanoseconds() > 0, true)
	require.Equal(r.AST, "{\"root\":{\"key\":\"val\"}}")

	err = d.Stop()
	require.NoError(err)
}

func TestDriverParserVersion(t *testing.T) {
	require := require.New(t)
	NativeBinary = "internal/native/mock"

	d, err := NewDriver(&uast.ObjectToNode{}, nil)
	require.NoError(err)
	require.NotNil(d)

	v := d.Version(nil)
	require.Equal(v.Version, "42")
	require.Equal(v.Build.String(), "2015-10-21 04:29:00 +0000 UTC")
}

func TestDriverDetails(t *testing.T) {
	require := require.New(t)

	manifestfeatures := []manifest.Feature{manifest.AST, manifest.UAST}
	expectedFeatures := []string{string(manifest.AST), string(manifest.UAST)}

	manifest := manifest.Manifest{
		Name:     "Foo",
		Language: "foo",
		Version:  "v0.1",
		Status:   manifest.Alpha,
		Features: manifestfeatures,
	}

	details := protocol.NewDriverDetails(&manifest)
	require.Equal(manifest.Name, details.Name)
	require.Equal(manifest.Language, details.Language)
	require.Equal(manifest.Version, details.Version)
	require.Equal(string(manifest.Status), details.Status)
	require.Equal(expectedFeatures, details.Features)
}
