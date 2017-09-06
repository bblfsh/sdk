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

	"gopkg.in/bblfsh/sdk.v0/uast"
)

//go:generate proteus  -f $GOPATH/src -p gopkg.in/bblfsh/sdk.v0/protocol -p gopkg.in/bblfsh/sdk.v0/uast

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

// ParseRequest is a request to parse a file and get its UAST.
//proteus:generate
type ParseRequest struct {
	// Filename is the name of the file containing the source code. Used for
	// language detection. Only filename is required, path might be used but
	// ignored. This is optional.
	Filename string
	// Language. If specified, it will override language detection. This is
	// optional.
	Language string
	// Content is the source code to be parsed.
	Content string `json:"content"`
	// Encoding is the encoding that the Content uses. Currently only UTF-8 and Base64
	// are supported.
	Encoding Encoding `json:"encoding"`
}

// ParseResponse is the reply to ParseRequest.
//proteus:generate
type ParseResponse struct {
	// Status is the status of the parsing request.
	Status Status `json:"status"`
	// Errors contains a list of parsing errors. If Status is ok, this list
	// should always be empty.
	Errors []string `json:"errors"`
	// UAST contains the UAST from the parsed code.
	UAST *uast.Node `json:"uast"`
}

// Parser can parse code to UAST.
type Parser interface {
	Parse(*ParseRequest) *ParseResponse
}

// DefaultParser is the default parser used by Parse and ParseNative.
var DefaultParser Parser

// Parse uses DefaultParser to process the given parsing request to get the UAST.
//proteus:generate
func Parse(req *ParseRequest) *ParseResponse {
	if DefaultParser == nil {
		return &ParseResponse{
			Status: Fatal,
			Errors: []string{"no default parser registered"},
		}
	}

	return DefaultParser.Parse(req)
}

// VersionRequest is a request to get server version
//proteus:generate
type VersionRequest struct {
}

// VersionResponse is the reply to VersionRequest
//proteus:generate
type VersionResponse struct {
	// Version is the server version
	Version string `json:"version"`
}

type Versioner interface {
	Version(*VersionRequest) *VersionResponse
}

// DefaultVersioner is the default versioner user by Version
var DefaultVersioner Versioner

// Version uses DefaultVersioner to process the given version request to get the version.
//proteus:generate
func Version(req *VersionRequest) *VersionResponse {
	if DefaultVersioner == nil {
		return &VersionResponse{
			Version: "unknown",
		}
	}

	return DefaultVersioner.Version(req)
}
