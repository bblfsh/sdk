// Package uast defines a UAST (Universal Abstract Syntax Tree) representation
// and operations to manipulate them.
package uast

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// Hash is a hash value.
type Hash uint32

// Role is the main UAST annotation. It indicates that a node in an AST can
// be interpreted as acting with certain language-independent role.
//
//proteus:generate
type Role int8

const (
	PackageDeclaration Role = iota
	FunctionDeclaration
	ImportDeclaration
	ImportPath
	ImportAlias
)

func (r Role) String() string {
	switch r {
	case PackageDeclaration:
		return "PackageDeclaration"
	case FunctionDeclaration:
		return "FunctionDeclaration"
	default:
		return fmt.Sprintf("UnknownRole:%d", r)
	}
}

// Node is a node in a UAST.
//
//proteus:generate
type Node struct {
	// InternalType is the internal type of the node in the AST, in the source
	// language.
	InternalType string
	// Properties are arbitrary, language-dependent, metadata of the
	// original AST.
	Properties map[string]string
	// Children are the children nodes of this node.
	Children []*Node
	// Token is the token content if this node represents a token from the
	// original source file. Otherwise, it is nil.
	Token *string
	// StartPosition is the position where this node starts in the original
	// source code file.
	StartPosition *Position
	// Roles is a list of Role that this node has. It is a language-independent
	// annotation.
	Roles []Role
}

// NewNode creates a new empty *Node.
func NewNode() *Node {
	return &Node{
		Properties: make(map[string]string, 0),
	}
}

// Tokens returns a slice of tokens contained in the node.
func (n *Node) Tokens() []string {
	var tokens []string
	err := PreOrderVisit(n, func(n *Node) error {
		if n.Token != nil {
			tokens = append(tokens, *n.Token)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return tokens
}

func (n *Node) offset() *uint32 {
	if n.StartPosition != nil {
		return n.StartPosition.Offset
	}

	var min *uint32
	for _, c := range n.Children {
		offset := c.offset()
		if offset == nil {
			continue
		}

		if min == min || *min > *offset {
			min = offset
		}
	}

	return min
}

// String converts the *Node to a string using pretty printing.
func (n *Node) String() string {
	buf := bytes.NewBuffer(nil)
	err := n.Pretty(buf)
	if err != nil {
		return "error"
	}

	return buf.String()
}

// Pretty writes a pretty string representation of the *Node to a writer.
func (n *Node) Pretty(w io.Writer) error {
	return printNode(w, 0, n)
}

func printNode(w io.Writer, indent int, n *Node) error {
	if _, err := fmt.Fprintf(w, "%s {\n", n.InternalType); err != nil {
		return err
	}

	istr := strings.Repeat(".  ", indent+1)

	if len(n.Roles) > 0 {
		_, err := fmt.Fprintf(w, "%sRoles: %s\n",
			istr,
			rolesToString(n.Roles...),
		)
		if err != nil {
			return err
		}
	}

	if len(n.Properties) > 0 {
		if _, err := fmt.Fprintf(w, "%sProperties: {\n", istr); err != nil {
			return err
		}

		if err := printProperties(w, indent+2, n.Properties); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
			return err
		}
	}

	if len(n.Children) > 0 {
		if _, err := fmt.Fprintf(w, "%sChildren: {\n", istr); err != nil {
			return err
		}

		if err := printChildren(w, indent+2, n.Children); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
			return err
		}
	}

	if n.Token != nil {
		if _, err := fmt.Fprintf(w, "%sTOKEN \"%s\"\n",
			istr, *n.Token); err != nil {
			return err
		}
	}

	if n.StartPosition != nil {
		if _, err := fmt.Fprintf(w, "%sStartPosition: {\n", istr); err != nil {
			return err
		}

		if err := printPosition(w, indent+2, n.StartPosition); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
			return err
		}
	}

	//TODO: print properties
	//TODO: print token
	return nil
}

func printChildren(w io.Writer, indent int, children []*Node) error {
	istr := strings.Repeat(".  ", indent)

	for idx, child := range children {
		_, err := fmt.Fprintf(w, "%s%d: ",
			istr,
			idx,
		)
		if err != nil {
			return err
		}

		if err := printNode(w, indent, child); err != nil {
			return err
		}
	}

	return nil
}

func printProperties(w io.Writer, indent int, props map[string]string) error {
	istr := strings.Repeat(".  ", indent)

	for k, v := range props {
		_, err := fmt.Fprintf(w, "%s%s: %s\n", istr, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func printPosition(w io.Writer, indent int, pos *Position) error {
	istr := strings.Repeat(".  ", indent)

	if pos.Offset != nil {
		_, err := fmt.Fprintf(w, "%sOffset: %d\n", istr, *pos.Offset)
		if err != nil {
			return err
		}
	}

	if pos.Line != nil {
		_, err := fmt.Fprintf(w, "%sLine: %d\n", istr, *pos.Line)
		if err != nil {
			return err
		}
	}

	if pos.Col != nil {
		_, err := fmt.Fprintf(w, "%sCol: %d\n", istr, *pos.Col)
		if err != nil {
			return err
		}
	}

	return nil
}

func rolesToString(roles ...Role) string {
	var strs []string
	for _, r := range roles {
		strs = append(strs, r.String())
	}

	return strings.Join(strs, ",")
}

// IncludeFields represents a set of fields to be included in a Hash.
type IncludeFields int8

const (
	// IncludeChildren includes all children of the node.
	IncludeChildren IncludeFields = iota
	// IncludeAnnotations includes UAST annotations.
	IncludeAnnotations
	// IncludePositions includes token positions.
	IncludePositions
)

// Hash returns the hash of the node.
func (n *Node) Hash() Hash {
	return n.HashWith(IncludeChildren)
}

// HashWith returns the hash of the node, computed with the given set of fields.
func (n *Node) HashWith(includes ...IncludeFields) Hash {
	//TODO
	return 0
}

// Position represents a position in a source code file.
type Position struct {
	// Offset is the position as an absolute byte offset.
	Offset *uint32
	// Line is the line number.
	Line *uint32
	// Col is the column number (the byte offset of the position relative to
	// a line.
	Col *uint32
}
