package driver

import (
	"io"

	"github.com/bblfsh/sdk/protocol"
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
