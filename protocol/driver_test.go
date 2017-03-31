package protocol

import (
	"bytes"
	"testing"

	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"

	"github.com/stretchr/testify/require"
)

func testDriver(t *testing.T, args []string, expectedErr bool, expectedStdout, expectedStderr string) {
	require := require.New(t)

	stdin := bytes.NewBuffer(nil)
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)

	d := &Driver{
		Version:  "test",
		Build:    "test",
		ToNoder:  &uast.BaseToNoder{},
		Annotate: ann.On(ann.Any).Roles(uast.SimpleIdentifier),
		In:       stdin,
		Out:      stdout,
		Err:      stderr,
	}

	err := d.run(args)
	if expectedErr {
		require.Error(err)
	} else {
		require.NoError(err)
	}
	require.Equal(expectedStdout, string(stdout.Bytes()))
	require.Equal(expectedStderr, string(stderr.Bytes()))
}

func TestDriverParseUAST(t *testing.T) {
	testDriver(
		t,
		[]string{
			"driver",
			"parse-uast",
			"--native-bin=internal/testnative/native",
			"./driver_test.go",
		},
		false,
		`{"status":"ok","errors":null,"uast":{"Properties":{"key":"0"},"StartPosition":{"Offset":0,"Line":0,"Col":0},"EndPosition":{"Offset":0,"Line":0,"Col":0},"Roles":[1]}}
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
			"--native-bin=internal/testnative/native",
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
  driver [OPTIONS] <parse-native | parse-uast | serve>

Help Options:
  -h, --help  Show this help message

Available commands:
  parse-native
  parse-uast
  serve

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
