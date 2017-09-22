package driver_test

import (
	"bytes"
	"testing"

	"gopkg.in/bblfsh/sdk.v1/etc/skeleton/driver/normalizer"
	. "gopkg.in/bblfsh/sdk.v1/protocol/driver"
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/ann"

	"github.com/stretchr/testify/require"
)

func testDriver(t *testing.T, args []string, expectedErr bool, expectedStdout, expectedStderr string) {
	require := require.New(t)

	stdin := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	d := &Driver{
		Version:       "test",
		Build:         "test",
		ParserBuilder: normalizer.ParserBuilder,
		Annotate:      ann.On(ann.Any).Roles(uast.Identifier),
		In:            stdin,
		Out:           stdout,
		Err:           stderr,
	}

	err := d.Run(args)
	if expectedErr {
		require.Error(err)
	} else {
		require.NoError(err)
	}
	require.Equal(expectedStdout, string(stdout.Bytes()))
	require.Equal(expectedStderr, string(stderr.Bytes()))
}

func TestDriverParse(t *testing.T) {
	testDriver(
		t,
		[]string{
			"driver",
			"parse",
			"--native-bin=../internal/testnative/native",
			"./driver_test.go",
		},
		false,
		`{"status":"ok","errors":null,"uast":{"Properties":{"key":"val"},"Roles":[1]}}
`,
		"",
	)
}

func TestDriverParseNative(t *testing.T) {
	testDriver(
		t,
		[]string{
			"driver",
			"parse-native",
			"--native-bin=../internal/testnative/native",
			"./driver_test.go",
		},
		false,
		`{"status":"ok","errors":null,"ast":{"root":{"key":"val"}}}
`,
		"",
	)
}

func TestDriverHelp(t *testing.T) {
	expectedStderr := ""
	expectedStdout := `Usage:
  driver [OPTIONS] <command>

Help Options:
  -h, --help  Show this help message

Available commands:
  docgen
  parse
  parse-native
  serve
  tokenize

Build information
  commit: test
  date:test
`

	testDriver(
		t,
		[]string{"driver", "--help"},
		false,
		expectedStdout,
		expectedStderr,
	)
}
