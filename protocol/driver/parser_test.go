package driver

import (
	"errors"
	"testing"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/uast"

	"github.com/stretchr/testify/require"
)

type mockASTParser struct {
	Response *protocol.ParseASTResponse
	Error    error
}

func (p *mockASTParser) ParseAST(req *protocol.ParseASTRequest) (*protocol.ParseASTResponse, error) {
	return p.Response, p.Error
}

func (p *mockASTParser) Close() error {
	return nil
}

type mockUASTParser struct {
	Response *protocol.ParseUASTResponse
	Error    error
}

func (p *mockUASTParser) ParseUAST(req *protocol.ParseUASTRequest) (*protocol.ParseUASTResponse, error) {
	return p.Response, p.Error
}

func (p *mockUASTParser) Close() error {
	return nil
}

func TestTransformationASTParserUnderlyingError(t *testing.T) {
	require := require.New(t)

	e := errors.New("test error")
	p := &mockASTParser{Error: e}
	tf := func(d []byte, n *uast.Node) error { return nil }
	tp := &TransformationASTParser{ASTParser: p, Transformation: tf}

	resp, err := tp.ParseAST(&protocol.ParseASTRequest{Content: "foo"})
	require.Equal(e, err)
	require.Nil(resp)
}

func TestTransformationASTParserTransformationError(t *testing.T) {
	require := require.New(t)

	e := errors.New("test error")
	p := &mockASTParser{Response: &protocol.ParseASTResponse{Status: protocol.Ok}}
	tf := func(d []byte, n *uast.Node) error { return e }
	tp := &TransformationASTParser{ASTParser: p, Transformation: tf}

	resp, err := tp.ParseAST(&protocol.ParseASTRequest{Content: "foo"})
	require.Equal(e, err)
	require.NotNil(resp)
}

func TestTransformationASTParser(t *testing.T) {
	require := require.New(t)

	p := &mockASTParser{Response: &protocol.ParseASTResponse{Status: protocol.Ok, AST: &uast.Node{}}}
	tf := func(d []byte, n *uast.Node) error {
		n.InternalType = "foo"
		return nil
	}
	tp := &TransformationASTParser{ASTParser: p, Transformation: tf}

	resp, err := tp.ParseAST(&protocol.ParseASTRequest{Content: "foo"})
	require.NoError(err)
	require.NotNil(resp)
	require.NotNil(resp.AST)
	require.Equal("foo", resp.AST.InternalType)
}
