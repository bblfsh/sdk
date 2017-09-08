package driver

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/protocol/jsonlines"
	"gopkg.in/bblfsh/sdk.v1/uast"

	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	require := require.New(t)

	inr, inw := io.Pipe()
	outr, outw := io.Pipe()

	client := NewClient(inw, outr)
	req := &protocol.ParseRequest{Content: "foo"}

	done := make(chan struct{})
	go func() {
		resp := client.Parse(req)
		require.Equal(protocol.Ok, resp.Status)
		require.Equal("bar", resp.UAST.InternalType)

		resp = client.Parse(req)
		require.Equal(protocol.Fatal, resp.Status)

		inw.Close()
		resp = client.Parse(req)
		require.Equal(protocol.Fatal, resp.Status)

		close(done)
	}()

	srvDec := jsonlines.NewDecoder(inr)
	srvEnc := jsonlines.NewEncoder(outw)
	srvReq := &protocol.ParseRequest{}
	err := srvDec.Decode(srvReq)
	require.NoError(err)
	require.Equal(req, srvReq)

	err = srvEnc.Encode(&protocol.ParseResponse{
		Status: protocol.Ok,
		UAST:   &uast.Node{InternalType: "bar"},
	})
	require.NoError(err)

	err = srvDec.Decode(srvReq)
	require.NoError(err)
	require.Equal(req, srvReq)
	fmt.Fprintln(outw, "GARGABE\"")

	<-done

	err = client.Close()
	require.NoError(err)
}

type nopWriteCloser struct {
	*bytes.Buffer
}

func (w *nopWriteCloser) Close() error { return nil }
