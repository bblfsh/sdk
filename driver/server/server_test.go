package server

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v2/protocol"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/driver/manifest"
	"gopkg.in/bblfsh/sdk.v2/driver/native"
)

func init() {
	ManifestLocation = "../native/internal/manifest.toml"
}

func newDriver(path string) (*service, error) {
	if path == "" {
		path = "../native/internal/mock"
	}
	m, err := manifest.Load(ManifestLocation)
	if err != nil {
		return nil, err
	}
	d, err := driver.NewDriverFrom(native.NewDriverAt(path, native.UTF8), m, driver.Transforms{})
	if err != nil {
		return nil, err
	}
	return &service{d: d}, nil
}

func TestDriverParserParse(t *testing.T) {
	require := require.New(t)

	d, err := newDriver("")
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
	require.Empty(r.Errors, "%v", r.Errors)
	require.Equal(protocol.Ok, r.Status)
	require.Equal("fixture", r.Language)
	require.Equal("foo.f", r.Filename)
	require.True(r.Elapsed.Nanoseconds() > 0)
	require.Equal(` {
.  Children: {
.  .  0:  {
.  .  .  Properties: {
.  .  .  .  internalRole: root
.  .  .  .  key: val
.  .  .  }
.  .  }
.  }
}
`, r.UAST.String())

	err = d.Stop()
	require.NoError(err)
}

func TestDriverParserParse_MissingLanguage(t *testing.T) {
	require := require.New(t)

	d, err := newDriver("")
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

	d, err := newDriver("echo")
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

	d, err := newDriver("")
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

	d, err := newDriver("")
	require.NoError(err)
	require.NotNil(d)

	v := d.Version(nil)
	require.Equal(v.Version, "42")
	require.Equal(v.Build.String(), "2015-10-21 04:29:00 +0000 UTC")
}
