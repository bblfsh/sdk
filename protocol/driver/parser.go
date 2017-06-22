package driver

import (
	"io"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"
)

type UASTParserOptions struct {
	NativeBin string `long:"native-bin" description:"alternative path for the native binary" default:"/opt/driver/bin/native"`
}

// UASTParserBuilder is a function that given ParserOptions creates a Parser.
type UASTParserBuilder func(UASTParserOptions) (UASTParser, error)

type UASTParser interface {
	io.Closer
	protocol.Parser
}

// TransformationUASTParser wraps another ASTParser and applies a transformation
// to its results.
type TransformationUASTParser struct {
	// UASTParser to delegate parsing.
	UASTParser
	// Transformation function to apply to resulting *uast.Node. The first
	// argument is the original source code from the request. Any
	// transformation to *uast.Node should be performed in-place. If error
	// is returned, it will be propagated to the ParseUAST call.
	Transformation func([]byte, *uast.Node) error
}

// ParseUAST calls the wrapped ASTParser and applies the transformation to its
// result.
func (p *TransformationUASTParser) ParseUAST(req *protocol.ParseUASTRequest) *protocol.ParseUASTResponse {
	resp := p.UASTParser.ParseUAST(req)
	if resp.Status == protocol.Fatal {
		return resp
	}

	if err := p.Transformation([]byte(req.Content), resp.UAST); err != nil {
		resp.Status = protocol.Error
		resp.Errors = append(resp.Errors, err.Error())
	}

	return resp
}

type annotationParser struct {
	UASTParser
	Annotation *ann.Rule
}

func (p *annotationParser) ParseUAST(req *protocol.ParseUASTRequest) *protocol.ParseUASTResponse {
	resp := p.UASTParser.ParseUAST(&protocol.ParseUASTRequest{
		Content:  req.Content,
		Encoding: req.Encoding,
	})
	if resp.Status == protocol.Fatal {
		return resp
	}

	if err := p.Annotation.Apply(resp.UAST); err != nil {
		resp.Errors = append(resp.Errors, err.Error())
	}

	return resp
}
