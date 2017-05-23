package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/bblfsh/sdk/uast"
)

//go:generate proteus  -f $GOPATH/src -p github.com/bblfsh/sdk/protocol -p github.com/bblfsh/sdk/uast

// Status is the status of a response.
//proteus:generate
type Status byte

const (
	// Ok status code.
	Ok Status = iota
	// Error status code. It is replied when the driver has got the AST with errors.
	Error
	// Fatal status code. It is replied when the driver hasn't could get the AST.
	Fatal
)

var statusToString = map[Status]string{
	Ok:    "ok",
	Error: "error",
	Fatal: "fatal",
}

var stringToStatus = map[string]Status{
	"ok":    Ok,
	"error": Error,
	"fatal": Fatal,
}

func (s Status) String() string {
	return statusToString[s]
}

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Status should be a string, got %s", data)
	}

	*s, _ = stringToStatus[str]
	return nil
}

// ParseUASTRequest is a request to parse a file and get its UAST.
//proteus:generate
type ParseUASTRequest struct {
	// Content is the source code to be parsed.
	Content string `json:"content"`
}

// ParseUASTResponse is the reply to ParseUASTRequest.
//proteus:generate
type ParseUASTResponse struct {
	// Status is the status of the parsing request.
	Status Status `json:"status"`
	// Errors contains a list of parsing errors. If Status is ok, this list
	// should always be empty.
	Errors []string `json:"errors"`
	// UAST contains the parsed UAST.
	UAST *uast.Node `json:"uast"`
}

// Parser can parse UAST.
type Parser interface {
	ParseUAST(*ParseUASTRequest) *ParseUASTResponse
}

// DefaultParser is the default parser used by ParseAST and ParseUAST.
var DefaultParser Parser

// ParseUAST uses DefaultParser to process the given UAST parsing request.
//proteus:generate
func ParseUAST(req *ParseUASTRequest) *ParseUASTResponse {
	if DefaultParser == nil {
		return &ParseUASTResponse{
			Status: Fatal,
			Errors: []string{"no default parser registered"},
		}
	}

	return DefaultParser.ParseUAST(req)
}
