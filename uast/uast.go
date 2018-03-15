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

// Hash is a hash value.
type Hash uint32

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

func AsPosition(m Object) Position {
	off, _ := m["off"].(Int)
	line, _ := m["line"].(Int)
	col, _ := m["col"].(Int)
	return Position{
		Offset: uint32(off),
		Line:   uint32(line),
		Col:    uint32(col),
	}
}

func (p Position) ToObject() Object {
	// TODO: add struct fields and generate this via reflection
	return Object{
		"off":  Int(p.Offset),
		"line": Int(p.Line),
		"col":  Int(p.Col),
	}
}

// IncludeFlag represents a set of fields to be included in a Hash or String.
type IncludeFlag int64

const (
	// IncludeChildren includes all children of the node.
	IncludeChildren IncludeFlag = 1
	// IncludeAnnotations includes UAST annotations.
	IncludeAnnotations = 2
	// IncludePositions includes token positions.
	IncludePositions = 4
	// IncludeTokens includes token contents.
	IncludeTokens = 8
	// IncludeInternalType includes internal type.
	IncludeInternalType = 16
	// IncludeProperties includes properties.
	IncludeProperties = 32
	// IncludeOriginalAST includes all properties that are present
	// in the original AST.
	IncludeOriginalAST = IncludeChildren |
		IncludePositions |
		IncludeTokens |
		IncludeInternalType |
		IncludeProperties
	// IncludeAll includes all fields.
	IncludeAll = IncludeOriginalAST | IncludeAnnotations
)

func (f IncludeFlag) Is(of IncludeFlag) bool {
	return f&of != 0
}

// Path represents a Node with its path in a tree. It is a slice with every
// token in the path, where the last one is the node itself. The empty path is
// is the zero value (e.g. parent of the root node).
type Path []*Node

// NewPath creates a new Path from a slice of nodes.
func NewPath(nodes ...*Node) Path {
	return Path(nodes)
}

// IsEmpty returns true if the path is empty.
func (p Path) IsEmpty() bool {
	return len(p) == 0
}

// Node returns the node. If the path is empty, the result is nil.
func (p Path) Node() *Node {
	if p.IsEmpty() {
		return nil
	}

	return p[len(p)-1]
}

// Child creates a Path for a given child.
func (p Path) Child(n *Node) Path {
	dst := make(Path, len(p)+1)
	copy(dst, p)
	dst[len(p)] = n
	return dst
}

// Parent returns the path of the parent of this node.
func (p Path) Parent() Path {
	if len(p) <= 1 {
		return Path(nil)
	}

	dst := make(Path, 0, len(p)-1)
	copy(dst, p)
	return dst
}

func Tokens(n *Node) []string {
	var tokens []string
	iter := NewOrderPathIter(NewPath(n))
	for {
		p := iter.Next()
		if p.IsEmpty() {
			break
		}

		n := p.Node()
		if n.Token != "" {
			tokens = append(tokens, n.Token)
		}
	}
	return tokens
}
