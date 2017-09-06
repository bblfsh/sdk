package driver

import (
	"io"

	"gopkg.in/bblfsh/sdk.v0/protocol"
	"gopkg.in/bblfsh/sdk.v0/uast"
	"gopkg.in/bblfsh/sdk.v0/uast/ann"
)

type ParserOptions struct {
	NativeBin string `long:"native-bin" description:"alternative path for the native binary" default:"/opt/driver/bin/native"`
}

// ParserBuilder is a function that given ParserOptions creates a Parser.
type ParserBuilder func(ParserOptions) (Parser, error)

type Parser interface {
	io.Closer
	protocol.Parser
}

// TransformationParser wraps another Parser and applies a transformation
// to its results.
type TransformationParser struct {
	// Parser to delegate parsing.
	Parser
	// Transformation function to apply to resulting *uast.Node. The first
	// argument is the original source code from the request. Any
	// transformation to *uast.Node should be performed in-place. If error
	// is returned, it will be propagated to the Parse call.
	Transformation func([]byte, *uast.Node) error
}

// Parse calls the wrapped Parser and applies the transformation to its
// result.
func (p *TransformationParser) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	resp := p.Parser.Parse(req)
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
	Parser
	Annotation *ann.Rule
}

func (p *annotationParser) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	resp := p.Parser.Parse(&protocol.ParseRequest{
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
