package driver

import (
	"errors"
	"testing"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/uast"

	"github.com/stretchr/testify/require"
)

type mockUASTParser struct {
	Response *protocol.ParseUASTResponse
}

func (p *mockUASTParser) ParseUAST(req *protocol.ParseUASTRequest) *protocol.ParseUASTResponse {
	return p.Response
}

func (p *mockUASTParser) Close() error {
	return nil
}

func TestTransformationUASTParserUnderlyingError(t *testing.T) {
	require := require.New(t)

	e := "test error"
	p := &mockUASTParser{Response: &protocol.ParseUASTResponse{Status: protocol.Fatal, Errors: []string{e}}}
	tf := func(d []byte, n *uast.Node) error { return nil }
	tp := &TransformationUASTParser{UASTParser: p, Transformation: tf}

	resp := tp.ParseUAST(&protocol.ParseUASTRequest{Content: "foo"})
	require.Equal(protocol.Fatal, resp.Status)
}

func TestTransformationUASTParserTransformationError(t *testing.T) {
	require := require.New(t)

	e := errors.New("test error")
	p := &mockUASTParser{Response: &protocol.ParseUASTResponse{Status: protocol.Ok}}
	tf := func(d []byte, n *uast.Node) error { return e }
	tp := &TransformationUASTParser{UASTParser: p, Transformation: tf}

	resp := tp.ParseUAST(&protocol.ParseUASTRequest{Content: "foo"})
	require.Equal(protocol.Error, resp.Status)
	require.Equal([]string{e.Error()}, resp.Errors)
}

func TestTransformationUASTParser(t *testing.T) {
	require := require.New(t)

	p := &mockUASTParser{Response: &protocol.ParseUASTResponse{Status: protocol.Ok, UAST: &uast.Node{}}}
	tf := func(d []byte, n *uast.Node) error {
		n.InternalType = "foo"
		return nil
	}
	tp := &TransformationUASTParser{UASTParser: p, Transformation: tf}

	resp := tp.ParseUAST(&protocol.ParseUASTRequest{Content: "foo"})
	require.NotNil(resp)
	require.NotNil(resp.UAST)
	require.Equal("foo", resp.UAST.InternalType)
}
