package uast

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

//proteus:generate
type Role int8

const (
	PackageDeclaration Role = iota
	FunctionDeclaration
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

//proteus:generate
type Node struct {
	// InternalType is the internal type of the node in the AST, in the source
	// language.
	InternalType string
	Roles        []Role
	Properties   map[string]string
	Children     []*Node
	Token        *Token
}

func NewNode() *Node {
	return &Node{
		Properties: make(map[string]string, 0),
	}
}

func (n *Node) String() string {
	buf := bytes.NewBuffer(nil)
	err := n.Pretty(buf)
	if err != nil {
		return "error"
	}

	return buf.String()
}

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

func rolesToString(roles ...Role) string {
	var strs []string
	for _, r := range roles {
		strs = append(strs, r.String())
	}

	return strings.Join(strs, ",")
}

//proteus:generate
type Token struct {
	Position *Position
	Content  string
}

type Position struct {
	Offset *uint32
	Line   *uint32
	Col    *uint32
}

type IncludeFields int8

const (
	IncludeChildren IncludeFields = iota
	IncludeAnnotations
	IncludePositions
)

type Hash uint32

func (n *Node) Hash() Hash {
	return n.HashWith(IncludeChildren)
}

func (n *Node) HashWith(includes ...IncludeFields) Hash {
	//TODO
	return 0
}
