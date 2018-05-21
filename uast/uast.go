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

// Package uast defines a UAST (Universal Abstract Syntax Tree) representation
// and operations to manipulate them.
package uast

import (
	"gopkg.in/bblfsh/sdk.v2/uast/role"
)

const (
	// TypePosition is a node type for positional information in AST. See AsPosition.
	TypePosition = "ast:Position"
	// TypeOperator is a node type for an operator AST node. See Operator.
	TypeOperator = "ast:Operator"
	// KeyPosOff is a name for a Position object field that stores a bytes offset.
	KeyPosOff = "off"
	// KeyPosLine is a name for a Position object field that stores a source line.
	KeyPosLine = "line"
	// KeyPosCol is a name for a Position object field that stores a source column.
	KeyPosCol = "col"
)

// Position represents a position in a source code file.
type Position struct {
	// Offset is the position as an absolute byte offset. It is a 0-based
	// index.
	Offset uint32
	// Line is the line number. It is a 1-based index.
	Line uint32
	// Col is the column number (the byte offset of the position relative to
	// a line. It is a 1-based index.
	Col uint32
}

// AsPosition transforms a generic AST node to a Position object.
func AsPosition(m Object) *Position {
	if m.Type() != TypePosition {
		return nil
	}
	off, _ := m[KeyPosOff].(Int)
	line, _ := m[KeyPosLine].(Int)
	col, _ := m[KeyPosCol].(Int)
	return &Position{
		Offset: uint32(off),
		Line:   uint32(line),
		Col:    uint32(col),
	}
}

// ToObject converts Position to a generic AST node.
func (p Position) ToObject() Object {
	// TODO: add struct fields and generate this via reflection
	return Object{
		KeyType:    String(TypePosition),
		KeyPosOff:  Int(p.Offset),
		KeyPosLine: Int(p.Line),
		KeyPosCol:  Int(p.Col),
	}
}

// RoleList converts a set of roles into a list node.
func RoleList(roles ...role.Role) Array {
	arr := make(Array, 0, len(roles))
	for _, r := range roles {
		arr = append(arr, String(r.String()))
	}
	return arr
}

// Tokens collects all tokens of the tree recursively (pre-order).
func Tokens(n Node) []string {
	var tokens []string
	WalkPreOrder(n, func(n Node) bool {
		if obj, ok := n.(Object); ok {
			if tok := obj.Token(); tok != "" {
				tokens = append(tokens, tok)
			}
		}
		return true
	})
	return tokens
}
