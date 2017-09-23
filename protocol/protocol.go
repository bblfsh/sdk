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

//go:generate proteus  -f $GOPATH/src -p gopkg.in/bblfsh/sdk.v1/protocol -p gopkg.in/bblfsh/sdk.v1/uast
//go:generate stringer -type=Status,Encoding -output stringer.go

package protocol

import (
	"time"

	"gopkg.in/bblfsh/sdk.v1/uast"
)

// Service can parse code to UAST or AST.
type Service interface {
	Parse(*ParseRequest) *ParseResponse
	NativeParse(*NativeParseRequest) *NativeParseResponse
	Version(*VersionRequest) *VersionResponse
}

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
	// Encoding is the encoding that the Content uses. Currently only UTF-8 and
	// Base64 are supported.
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
	// Elapsed is the amount of time consume processing the request.
	Elapsed time.Duration `json:"elapsed"`
}

// NativeParseRequest is a request to parse a file and get its native AST.
//proteus:generate
type NativeParseRequest ParseRequest

// NativeParseResponse is the reply to NativeParseRequest by the native parser.
//proteus:generate
type NativeParseResponse struct {
	// Status is the status of the parsing request.
	Status Status `json:"status"`
	// Status is the status of the parsing request.
	Errors []string `json:"errors"`
	// AST contains the AST from the parsed code.
	AST interface{} `json:"ast"`
	// Elapsed is the amount of time consume processing the request.
	Elapsed time.Duration `json:"elapsed"`
}

// DefaultParser is the default parser used by Parse and ParseNative.
var DefaultParser Service

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
type VersionRequest struct{}

// VersionResponse is the reply to VersionRequest
//proteus:generate
type VersionResponse struct {
	// Version is the server version.
	Version string `json:"version"`
	// Build contains the timestamp at the time of the build.
	Build string `json:"build"`
	// Commit used to compile this code, the commit will contain a `+` at the
	// end of hash when the repository contained uncommitted changes.
	Commit string `json:"commit"`
}

// DefaultVersioner is the default versioner user by Version
var DefaultVersioner Service

// Version uses DefaultVersioner to process the given version request to get the version.
//proteus:generate
func Version(req *VersionRequest) *VersionResponse {
	const noset = "no-set"

	if DefaultVersioner == nil {
		return &VersionResponse{
			Version: noset,
			Build:   noset,
			Commit:  noset,
		}
	}

	return DefaultVersioner.Version(req)
}
