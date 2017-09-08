package driver

import (
	"errors"
	"testing"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"

	"github.com/stretchr/testify/require"
)

type mockParser struct {
	Response *protocol.ParseResponse
}

func (p *mockParser) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	return p.Response
}

func (p *mockParser) Close() error {
	return nil
}

func TestTransformationParserUnderlyingError(t *testing.T) {
	require := require.New(t)

	e := "test error"
	p := &mockParser{Response: &protocol.ParseResponse{Status: protocol.Fatal, Errors: []string{e}}}
	tf := func(d []byte, n *uast.Node) error { return nil }
	tp := &TransformationParser{Parser: p, Transformation: tf}

	resp := tp.Parse(&protocol.ParseRequest{Content: "foo"})
	require.Equal(protocol.Fatal, resp.Status)
}

func TestTransformationParserTransformationError(t *testing.T) {
	require := require.New(t)

	e := errors.New("test error")
	p := &mockParser{Response: &protocol.ParseResponse{Status: protocol.Ok}}
	tf := func(d []byte, n *uast.Node) error { return e }
	tp := &TransformationParser{Parser: p, Transformation: tf}

	resp := tp.Parse(&protocol.ParseRequest{Content: "foo"})
	require.Equal(protocol.Error, resp.Status)
	require.Equal([]string{e.Error()}, resp.Errors)
}

func TestTransformationParser(t *testing.T) {
	require := require.New(t)

	p := &mockParser{Response: &protocol.ParseResponse{Status: protocol.Ok, UAST: &uast.Node{}}}
	tf := func(d []byte, n *uast.Node) error {
		n.InternalType = "foo"
		return nil
	}
	tp := &TransformationParser{Parser: p, Transformation: tf}

	resp := tp.Parse(&protocol.ParseRequest{Content: "foo"})
	require.NotNil(resp)
	require.NotNil(resp.UAST)
	require.Equal("foo", resp.UAST.InternalType)
}
