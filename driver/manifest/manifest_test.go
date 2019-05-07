package manifest

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fixture = `
name = "Foo"
language = "foo"
status = ""
features = ["ast", "uast", "roles"]

[documentation]
  description = "foo"
`[1:]

func TestEncode(t *testing.T) {
	m := &Manifest{}
	m.Name = "Foo"
	m.Language = "foo"
	m.Features = []Feature{AST, UAST, Roles}
	m.Documentation = &Documentation{
		Description: "foo",
	}

	buf := bytes.NewBuffer(nil)
	err := m.Encode(buf)
	assert.Nil(t, err)

	assert.Equal(t, fixture, buf.String())
}

func TestDecode(t *testing.T) {
	m := &Manifest{}

	buf := bytes.NewBufferString(fixture)
	err := m.Decode(buf)
	assert.Nil(t, err)

	assert.Equal(t, "foo", m.Language)
}

func TestCurrentSDKVersion(t *testing.T) {
	require.Equal(t, 3, CurrentSDKMajor())
}

func TestParseMaintainers(t *testing.T) {
	m := parseMaintainers(strings.NewReader(`
John Doe <john@domain.com> (@john_at_github)
Bob <bob@domain.com>
`))
	require.Equal(t, []Maintainer{
		{Name: "John Doe", Email: "john@domain.com", Github: "john_at_github"},
		{Name: "Bob", Email: "bob@domain.com"},
	}, m)
}

var casesVersion = []struct {
	name   string
	files  map[string]string
	expect string
}{
	{
		name:   "no files",
		expect: "1",
	},
	{
		name: "dep lock v1",
		files: map[string]string{
			"Gopkg.lock": `
[[projects]]
  name = "gopkg.in/bblfsh/sdk.v1"
  version = "v1.16.1"
`,
		},
		expect: "1.16.1",
	},
	{
		name: "dep lock both",
		files: map[string]string{
			"Gopkg.lock": `
[[projects]]
  name = "gopkg.in/bblfsh/sdk.v1"
  version = "v1.16.1"

[[projects]]
  name = "gopkg.in/bblfsh/sdk.v2"
  version = "v2.2.1"
`,
		},
		expect: "2.2.1",
	},
	{
		name: "dep lock no vers",
		files: map[string]string{
			"Gopkg.lock": `
[[projects]]
  name = "gopkg.in/bblfsh/sdk.v1"

[[projects]]
  name = "gopkg.in/bblfsh/sdk.v2"
`,
		},
		expect: "2",
	},
	{
		name: "dep toml x",
		files: map[string]string{
			"Gopkg.toml": `
[[constraint]]
  name = "gopkg.in/bblfsh/sdk.v1"
  version = "1.16.x"
`,
		},
		expect: "1.16",
	},
	{
		name: "go mod",
		files: map[string]string{
			"go.mod": `module github.com/bblfsh/some-driver

require (
	gopkg.in/bblfsh/sdk.v1 v1.16.1
	github.com/bblfsh/sdk/v3 v3.1.2
)
`,
		},
		expect: "3.1.2",
	},
	{
		name: "go mod legacy",
		files: map[string]string{
			"go.mod": `module github.com/bblfsh/some-driver

require (
	gopkg.in/bblfsh/sdk.v1 v1.16.1
)
`,
		},
		expect: "1.16.1",
	},
	{
		name: "go mod incompatible",
		files: map[string]string{
			"go.mod": `module github.com/bblfsh/some-driver

require (
	gopkg.in/bblfsh/sdk.v1 v1.16.1
	github.com/bblfsh/sdk/v3 v3.1.2+incompatible
)
`,
		},
		expect: "3.1.2",
	},
	{
		name: "go mod no tag",
		files: map[string]string{
			"go.mod": `module github.com/bblfsh/some-driver

require (
	gopkg.in/bblfsh/sdk.v1 v1.16.1
	github.com/bblfsh/sdk/v3 v3.0.0-20190326155454-bbb149502c30
)
`,
		},
		expect: "3.0.0",
	},
}

func TestSDKVersion(t *testing.T) {
	for _, c := range casesVersion {
		c := c
		t.Run(c.name, func(t *testing.T) {
			vers, err := SDKVersion(func(path string) ([]byte, error) {
				data, ok := c.files[path]
				if !ok {
					return nil, nil
				}
				return []byte(data), nil
			})
			require.NoError(t, err)
			require.Equal(t, c.expect, vers)
		})
	}
}
