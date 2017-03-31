package protocol_test

import (
	"bufio"
	"os/exec"
	"testing"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/jsonlines"

	"github.com/stretchr/testify/require"
)

func TestDriverMain(t *testing.T) {
	require := require.New(t)

	cmd := exec.Command(
		"go", "run", "internal/testdriver/main.go",
		"serve",
		"--native-bin", "internal/testnative/native",
	)

	in, err := cmd.StdinPipe()
	require.NoError(err)

	out, err := cmd.StdoutPipe()
	require.NoError(err)

	errOut, err := cmd.StderrPipe()
	require.NoError(err)

	err = cmd.Start()
	require.NoError(err)

	go func() {
		buf := bufio.NewReader(errOut)
		for {
			line, _, err := buf.ReadLine()
			t.Log("ERR:", string(line))
			if err != nil {
				break
			}
		}
	}()

	enc := jsonlines.NewEncoder(in)
	dec := jsonlines.NewDecoder(out)

	req := &protocol.ParseUASTRequest{Content: "foo"}
	err = enc.Encode(req)
	require.NoError(err)

	resp := &protocol.ParseUASTResponse{}
	err = dec.Decode(resp)
	require.NoError(err)
	require.Equal(protocol.Ok, resp.Status)

	err = in.Close()
	require.NoError(err)

	err = cmd.Wait()
	require.NoError(err)
}
