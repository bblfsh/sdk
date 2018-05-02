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

// go:generate proteus  -f $GOPATH/src -p gopkg.in/bblfsh/sdk.v1/protocol -p gopkg.in/bblfsh/sdk.v1/uast -p gopkg.in/bblfsh/sdk.v1/uast/role
//go:generate protoc --proto_path=$GOPATH/src:. --gogo_out=plugins=grpc:. ./generated.proto
//go:generate stringer -type=Status,Encoding -output stringer.go

package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

// DefaultService is the default service used to process requests.
var DefaultService Service

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

// Response basic response, never used directly.
type Response struct {
	// Status is the status of the parsing request.
	Status Status `json:"status"`
	// Status is the status of the parsing request.
	Errors []string `json:"errors"`
	// Elapsed is the amount of time consume processing the request.
	Elapsed time.Duration `json:"elapsed"`
}

// ParseRequest is a request to parse a file and get its UAST.
//proteus:generate
type ParseRequest struct {
	// Filename is the name of the file containing the source code. Used for
	// language detection. Only filename is required, path might be used but
	// ignored. This is optional.
	Filename string `json:"filename"`
	// Language. If specified, it will override language detection. This is
	// optional.
	Language string `json:"language"`
	// Content is the source code to be parsed.
	Content string `json:"content"`
	// Encoding is the encoding that the Content uses. Currently only UTF-8 and
	// Base64 are supported.
	Encoding Encoding `json:"encoding"`
	// Timeout amount of time for wait until the request is proccessed.
	Timeout time.Duration `json:"timeout"`
}

// ParseResponse is the reply to ParseRequest.
//proteus:generate
type ParseResponse struct {
	Response
	// UAST contains the UAST from the parsed code.
	UAST *Node `json:"uast"`
	// Language. The language that was parsed. Usedful if you used language
	// autodetection for the request.
	Language string `json:"language"`
	// Filename is the name of the file containing the source code. Used for
	// language detection. Only filename is required, path might be used but
	// ignored. This is optional.
	Filename string `json:"filename"`
}

func (r *ParseResponse) String() string {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, "Status: ", strings.ToLower(r.Status.String()))
	fmt.Fprintln(buf, "Language: ", strings.ToLower(r.Language))
	if len(r.Filename) > 0 {
		fmt.Fprintln(buf, "Filename:: ", strings.ToLower(r.Filename))
	}
	fmt.Fprintln(buf, "Errors: ")
	for _, err := range r.Errors {
		fmt.Fprintln(buf, " . ", err)
	}

	fmt.Fprintln(buf, "UAST: ")
	fmt.Fprintln(buf, r.UAST.String())

	return buf.String()
}

// NativeParseRequest is a request to parse a file and get its native AST.
//proteus:generate
type NativeParseRequest ParseRequest

// NativeParseResponse is the reply to NativeParseRequest by the native parser.
//proteus:generate
type NativeParseResponse struct {
	Response
	// AST contains the AST from the parsed code in json format.
	AST string `json:"ast"`
	// Language. The language that was parsed. Usedful if you used language
	// autodetection for the request.
	Language string `json:"language"`
}

func (r *NativeParseResponse) String() string {
	var s struct {
		Status   string      `json:"status"`
		Language string      `json:"language"`
		Errors   []string    `json:"errors"`
		AST      interface{} `json:"ast"`
	}

	s.Status = strings.ToLower(r.Status.String())
	s.Language = strings.ToLower(r.Language)
	s.Errors = r.Errors
	if len(s.Errors) == 0 {
		s.Errors = make([]string, 0)
	}

	if len(r.AST) > 0 {
		err := json.Unmarshal([]byte(r.AST), &s.AST)
		if err != nil {
			return err.Error()
		}
	}

	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.SetIndent("", "    ")
	e.SetEscapeHTML(false)

	err := e.Encode(s)
	if err != nil {
		return err.Error()
	}

	return buf.String()
}

// VersionRequest is a request to get server version
//proteus:generate
type VersionRequest struct{}

// VersionResponse is the reply to VersionRequest
//proteus:generate
type VersionResponse struct {
	Response
	// Version is the server version. If is a local compilation the version
	// follows the pattern dev-<short-commit>[-dirty], dirty means that was
	// compile from a repository with un-committed changes.
	Version string `json:"version"`
	// Build contains the timestamp at the time of the build.
	Build time.Time `json:"build"`
}

// Node is a node in a UAST.
//
//proteus:generate
type Node struct {
	// InternalType is the internal type of the node in the AST, in the source
	// language.
	InternalType string `json:",omitempty"`
	// Properties are arbitrary, language-dependent, metadata of the
	// original AST.
	Properties map[string]string `json:",omitempty"`
	// Children are the children nodes of this node.
	Children []*Node `json:",omitempty"`
	// Token is the token content if this node represents a token from the
	// original source file. If it is empty, there is no token attached.
	Token string `json:",omitempty"`
	// StartPosition is the position where this node starts in the original
	// source code file.
	StartPosition *uast.Position `json:",omitempty"`
	// EndPosition is the position where this node ends in the original
	// source code file.
	EndPosition *uast.Position `json:",omitempty"`
	// Roles is a list of Role that this node has. It is a language-independent
	// annotation.
	Roles []role.Role `json:",omitempty"`
}

// NewNode creates a new empty *Node.
func NewNode() *Node {
	return &Node{
		Properties: make(map[string]string, 0),
		Roles:      []role.Role{role.Unannotated},
	}
}

// String converts the *Node to a string using pretty printing.
func (n *Node) String() string {
	buf := bytes.NewBuffer(nil)
	err := Pretty(n, buf, IncludeAll)
	if err != nil {
		return "error"
	}

	return buf.String()
}

const (
	// InternalRoleKey is a key string uses in properties to use the internal
	// role of a node in the AST, if any.
	InternalRoleKey = "internalRole"
)

// ToNode converts a generic AST node to Node object used in the protocol.
func ToNode(n uast.Node) (*Node, error) {
	nd, err := asNode(n, "")
	if err != nil {
		return nil, err
	}
	switch len(nd) {
	case 0:
		return nil, nil
	case 1:
		return nd[0], nil
	default:
		return &Node{Children: nd}, nil
	}
}

func arrayAsNode(n uast.Array, field string) ([]*Node, error) {
	arr := make([]*Node, 0, len(n))
	for _, s := range n {
		nd, err := asNode(s, field)
		if err != nil {
			return arr, err
		}
		arr = append(arr, nd...)
	}
	return arr, nil
}

func objectAsNode(n uast.Object, field string) ([]*Node, error) {
	nd := &Node{
		InternalType:  n.Type(),
		Token:         n.Token(),
		Roles:         n.Roles(),
		StartPosition: n.StartPosition(),
		EndPosition:   n.EndPosition(),
		Properties:    make(map[string]string),
	}
	if field != "" {
		nd.Properties[InternalRoleKey] = field
	}

	for k, v := range n {
		switch k {
		case uast.KeyType, uast.KeyToken, uast.KeyRoles,
			uast.KeyStart, uast.KeyEnd:
			// already processed
			continue
		}
		if nv, ok := v.(uast.Value); ok {
			nd.Properties[k] = fmt.Sprint(nv.Native())
		} else {
			sn, err := asNode(v, k)
			if err != nil {
				return nil, err
			}
			nd.Children = append(nd.Children, sn...)
		}
	}
	sort.Stable(byOffset(nd.Children))
	return []*Node{nd}, nil
}

func valueAsNode(n uast.Value, field string) ([]*Node, error) {
	nd := &Node{
		Token:      fmt.Sprint(n),
		Properties: make(map[string]string),
	}
	if field != "" {
		nd.Properties[InternalRoleKey] = field
	}
	return []*Node{nd}, nil
}

func asNode(n uast.Node, field string) ([]*Node, error) {
	switch n := n.(type) {
	case nil:
		return nil, nil
	case uast.Array:
		return arrayAsNode(n, field)
	case uast.Object:
		return objectAsNode(n, field)
	case uast.Value:
		return valueAsNode(n, field)
	default:
		return nil, fmt.Errorf("argument should be a node or a list, got: %T", n)
	}
}

type byOffset []*Node

func (s byOffset) Len() int      { return len(s) }
func (s byOffset) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byOffset) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	apos := startPosition(a)
	bpos := startPosition(b)
	if apos != nil && bpos != nil {
		if apos.Offset != bpos.Offset {
			return apos.Offset < bpos.Offset
		}
	} else if (apos == nil && bpos != nil) || (apos != nil && bpos == nil) {
		return bpos != nil
	}
	field1, ok1 := a.Properties[InternalRoleKey]
	field2, ok2 := b.Properties[InternalRoleKey]
	if ok1 && ok2 {
		return field1 < field2
	}
	return false
}

func startPosition(n *Node) *uast.Position {
	if n.StartPosition != nil {
		return n.StartPosition
	}

	var min *uast.Position
	for _, c := range n.Children {
		other := startPosition(c)
		if other == nil {
			continue
		}

		if min == nil || other.Offset < min.Offset {
			min = other
		}
	}

	return min
}
