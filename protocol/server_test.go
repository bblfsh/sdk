package protocol

import (
	"io"
	"testing"
	"time"

	"github.com/bblfsh/sdk/protocol/jsonlines"
	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"

	"github.com/stretchr/testify/require"
)

func TestServerOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		err := enc.Encode(&ParseUASTRequest{Content: "foo"})
		require.NoError(err)

		resp := &ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func TestServerOneMalformedAndOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		err := enc.Encode("BAD REQUEST")
		require.NoError(err)

		resp := &ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)
		require.Equal(Fatal, resp.Status)

		err = enc.Encode(&ParseUASTRequest{Content: "foo"})
		require.NoError(err)

		resp = &ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func testServer(t *testing.T, exitError bool, f func(io.WriteCloser, io.Reader)) {
	require := require.New(t)

	sIn, cIn := io.Pipe()
	cOut, sOut := io.Pipe()

	n, err := testExecNative()
	require.NoError(err)
	require.NotNil(n)

	s := &Server{
		In:       sIn,
		Out:      sOut,
		Native:   n,
		ToNoder:  &uast.BaseToNoder{},
		Annotate: ann.On(ann.Any),
	}

	err = s.Start()
	require.NoError(err)

	f(cIn, cOut)

	waitDone := make(chan struct{})
	go func() {
		err = s.Wait()
		if exitError {
			require.Error(err)
		} else {
			require.NoError(err)
		}
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.NewTicker(5 * time.Second).C:
		require.FailNow("wait timed out")
	}
}
