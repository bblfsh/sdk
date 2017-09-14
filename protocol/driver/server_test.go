package driver

import (
	"io"
	"testing"
	"time"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/protocol/jsonlines"

	"github.com/stretchr/testify/require"
)

func TestServerOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(p *mockParser, in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		p.Response = &protocol.ParseResponse{}
		err := enc.Encode(&protocol.ParseRequest{Content: "foo"})
		require.NoError(err)

		resp := &protocol.ParseResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func TestServerOneMalformedAndOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(p *mockParser, in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		p.Response = &protocol.ParseResponse{}
		err := enc.Encode("BAD REQUEST")
		require.NoError(err)

		resp := &protocol.ParseResponse{}
		err = dec.Decode(resp)
		require.NoError(err)
		require.Equal(protocol.Fatal, resp.Status)

		err = enc.Encode(&protocol.ParseRequest{Content: "foo"})
		require.NoError(err)

		resp = &protocol.ParseResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func TestServerOneFatalAndOneGood(t *testing.T) {
	require := require.New(t)
	testServer(t, false, func(p *mockParser, in io.WriteCloser, out io.Reader) {
		enc := jsonlines.NewEncoder(in)
		dec := jsonlines.NewDecoder(out)

		p.Response = &protocol.ParseResponse{Status: protocol.Fatal}
		err := enc.Encode(&protocol.ParseRequest{Content: "FATAL"})
		require.NoError(err)

		resp := &protocol.ParseResponse{}
		err = dec.Decode(resp)
		require.NoError(err)
		require.Equal(protocol.Fatal, resp.Status)

		p.Response = &protocol.ParseResponse{}
		err = enc.Encode(&protocol.ParseRequest{Content: "foo"})
		require.NoError(err)

		resp = &protocol.ParseResponse{}
		err = dec.Decode(resp)
		require.NoError(err)

		require.NoError(in.Close())
	})
}

func testServer(t *testing.T, exitError bool, f func(*mockParser, io.WriteCloser, io.Reader)) {
	require := require.New(t)

	sIn, cIn := io.Pipe()
	cOut, sOut := io.Pipe()

	p := &mockParser{}
	s := &Server{
		In:     sIn,
		Out:    sOut,
		Parser: p,
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
