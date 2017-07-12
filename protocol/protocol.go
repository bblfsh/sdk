// Copyright 2017 Sourced Technologies SL
// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy
// of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

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

// Encoding is the encoding used for the content string. Currently only
// UTF-8 or Base64 encodings are supported. You should use UTF-8 if you can
// and Base64 as a fallback.
//proteus:generate
type Encoding byte

const (
	// UTF8 encoding
	UTF8 Encoding = iota
	// Base64 encoding
	Base64
)

var encodingToString = map[Encoding]string{
	UTF8:   "UTF8",
	Base64: "Base64",
}

var stringToEncoding = map[string]Encoding{
	"UTF8":   UTF8,
	"Base64": Base64,
}

func (e Encoding) String() string {
	return encodingToString[e]
}

func (e Encoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Encoding should be a string, got %s", data)
	}

	*e, _ = stringToEncoding[str]
	return nil
}

// ParseUASTRequest is a request to parse a file and get its UAST.
//proteus:generate
type ParseUASTRequest struct {
	// Path is the path of the file containing the source code. Used for
	// language detection. This is optional.
	Path string
	// Language. If specified, it will override language detection. This is
	// optional.
	Language string
	// Content is the source code to be parsed.
	Content string `json:"content"`
	// Encoding is the encoding that the Content uses. Currently only UTF-8 and Base64
	// are supported.
	Encoding Encoding `json:"encoding"`
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

// DefaultParser is the default parser used by Parse and ParseNative.
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
