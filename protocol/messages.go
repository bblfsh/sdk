package protocol

import (
	"github.com/bblfsh/sdk/uast"
)

// Status is the status of a response.
type Status string

const (
	// Ok status code.
	Ok Status = "ok"
	// Error status code. It is replied when the driver has got the AST with errors.
	Error = "error"
	// Fatal status code. It is replied when the driver hasn't could get the AST.
	Fatal = "fatal"
)

// String returns the string value of the Status.
func (s Status) String() string {
	return string(s)
}

// ParseUASTRequest is a request to parse a file and get its UAST.
//proteus:generate
type ParseUASTRequest struct {
	// Content is the source code to be parsed.
	Content string
}

// ParseUASTResponse is the reply to ParseUASTRequest.
//proteus:generate
type ParseUASTResponse struct {
	// Status is the status of the parsing request.
	Status Status
	// Errors contrains a list of parsing errors. If Status is ok, this list
	// should always be empty.
	Errors []string
	// UAST contains the parsed UAST.
	UAST *uast.Node
}

// ParseASTRequest is a request to parse a file and get its AST.
//proteus:generate
type ParseASTRequest struct {
	// Content is the source code to be parsed.
	Content string
}

// ParseASTResponse is the reply to ParseASTRequest.
//proteus:generate
type ParseASTResponse struct {
	// Status is the status of the parsing request.
	Status Status
	// Errors contrains a list of parsing errors. If Status is ok, this list
	// should always be empty.
	Errors []string
	// AST contains the parsed AST in its normalized form. That is the same
	// type as the UAST, but without any annotation.
	AST *uast.Node
}
