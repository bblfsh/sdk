package driver

import (
	"io"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"
)

type ASTParserOptions struct {
	NativeBin string `long:"native-bin" description:"alternative path for the native binary" default:"/opt/driver/bin/native"`
}

// ASTParserBuilder is a function that given ParserOptions creates a Parser.
type ASTParserBuilder func(ASTParserOptions) (ASTParser, error)

type ASTParser interface {
	io.Closer
	ParseAST(req *protocol.ParseASTRequest) (*protocol.ParseASTResponse, error)
}

type UASTParser interface {
	io.Closer
	ParseUAST(req *protocol.ParseUASTRequest) (*protocol.ParseUASTResponse, error)
}

// TransformationASTParser wraps another ASTParser and applies a transformation
// to its results.
type TransformationASTParser struct {
	// ASTParser to delegate parsing.
	ASTParser
	// Transformation function to apply to resulting *uast.Node. The first
	// argument is the original source code from the request. Any
	// transformation to *uast.Node should be performed in-place. If error
	// is returned, it will be propagated to the ParseAST call.
	Transformation func([]byte, *uast.Node) error
}

// ParseAST calls the wrapped ASTParser and applies the transformation to its
// result.
func (p *TransformationASTParser) ParseAST(req *protocol.ParseASTRequest) (*protocol.ParseASTResponse, error) {
	resp, err := p.ASTParser.ParseAST(req)
	if err != nil || resp.Status == protocol.Fatal {
		return resp, err
	}

	return resp, p.Transformation([]byte(req.Content), resp.AST)
}

type uastParser struct {
	ASTParser
	Annotation *ann.Rule
}

func (p *uastParser) ParseUAST(req *protocol.ParseUASTRequest) (*protocol.ParseUASTResponse, error) {
	astResp, err := p.ASTParser.ParseAST(&protocol.ParseASTRequest{
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}

	resp := &protocol.ParseUASTResponse{
		Status: astResp.Status,
		Errors: astResp.Errors,
		UAST:   astResp.AST,
	}

	if resp.Status == protocol.Fatal {
		return resp, nil
	}

	if err := p.Annotation.Apply(resp.UAST); err != nil {
		resp.Errors = append(resp.Errors, err.Error())
	}

	return resp, nil
}

func (p *uastParser) Close() error {
	return p.ASTParser.Close()
}
