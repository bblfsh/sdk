package driver

import (
	"io"
	"testing"
	"time"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/jsonlines"

	"github.com/stretchr/testify/require"
)

func TestServerOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(p *mockUASTParser, in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		p.Response = &protocol.ParseUASTResponse{}
		err := enc.Encode(&protocol.ParseUASTRequest{Content: "foo"})
		require.NoError(err)

		resp := &protocol.ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func TestServerOneMalformedAndOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(p *mockUASTParser, in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		p.Response = &protocol.ParseUASTResponse{}
		err := enc.Encode("BAD REQUEST")
		require.NoError(err)

		resp := &protocol.ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)
		require.Equal(protocol.Fatal, resp.Status)

		err = enc.Encode(&protocol.ParseUASTRequest{Content: "foo"})
		require.NoError(err)

		resp = &protocol.ParseUASTResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func testServer(t *testing.T, exitError bool, f func(*mockUASTParser, io.WriteCloser, io.Reader)) {
	require := require.New(t)

	sIn, cIn := io.Pipe()
	cOut, sOut := io.Pipe()

	p := &mockUASTParser{}
	s := &Server{
		In:         sIn,
		Out:        sOut,
		UASTParser: p,
	}

	err := s.Start()
	require.NoError(err)

	f(p, cIn, cOut)

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
