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
//go:generate stringer -type=Role
type Role int8

const (
	// SimpleIdentifier is the most basic form of identifier, used for variable
	// names, functions, packages, etc.
	SimpleIdentifier Role = iota
	// QualifiedIdentifier is form of identifier composed of multiple
	// SimpleIdentifier. One main identifier (usually the last one) and one
	// or more qualifiers.
	QualifiedIdentifier

	Expression
	Statement

	// File is the root node of a single file AST.
	File

	// PackageDeclaration identifies the package that all its children
	// belong to. Its children include, at least, QualifiedIdentifier or
	// SimpleIdentifier with the package name.
	PackageDeclaration

	// ImportDeclaration represents the import of another package in the
	// current scope. Its children may include an ImportPath and ImportInclude.
	ImportDeclaration
	// ImportPath is the (usually) fully qualified package name to import.
	ImportPath
	// ImportAlias is an identifier used as an alias for an imported package
	// in a certain scope.
	ImportAlias

	// TODO: arguments, return value, body, etc
	FunctionDeclaration

	// TypeDeclaration is the declaration of a type. It could be a class or
	// interface in Java, a struct, interface or alias in Go, etc.
	TypeDeclaration

	// If is used for if-then[-else] statements or expressions.
	// An if-then tree will look like:
	//
	// 	IfStatement {
	//		**[non-If nodes] {
	//			IfCondition {
	//				[...]
	//                      }
	//		}
	//		**[non-If* nodes] {
	//			IfBody {
	//				[...]
	//			}
	//		}
	//		**[non-If* nodes] {
	//			IfElse {
	//				[...]
	//			}
	//		}
	//	}
	//
	// The IfElse node is optional. The order of IfCondition, IfBody and
	// IfElse is not defined.
	If
	// IfCondition is a condition in an IfStatement or IfExpression.
	IfCondition
	// IfBody is the code following a then clause in an IfStatement or
	// IfExpression.
	IfBody
	// IfBody is the code following a else clause in an IfStatement or
	// IfExpression.
	IfElse

	// Switch is used to represent a broad of switch flavors. An expression
	// is evaluated and then compared to the values returned by different
	// case expressions, executing a body associated to the first case that
	// matches. Similar constructions that go beyond expression comparison
	// (such as pattern matching in Scala's match) should not be annotated
	// with Switch.
	//
	// TODO: We still have to decide how to annotate fallthrough and
	//      non-fallthrough variants. As well as crazy variants such as Perl
	//      and Bash with its optional fallthrough.
	Switch
	SwitchCase
	SwitchCaseCondition
	SwitchCaseBody
	SwitchDefault

	For
	ForInit
	ForExpression
	ForUpdate
	ForBody

	ForEach

	While
	WhileCondition
	WhileBody

	DoWhile
	DoWhileCondition
	DoWhileBody

	Break
	Continue

	// Block is a group of statements. If the source language has block scope,
	// it should be annotated both with Block and BlockScope.
	Block
	// BlockScope is a block with its own block scope.
	// TODO: Should we replace BlockScope with a more general Scope role that
	//       can be combined with Block?
	BlockScope

	// Return is a return statement. It might have a child expression or not
	// as with naked returns in Go or return in void methods in Java.
	Return

	Try
	TryBody
	TryCatch
	TryFinally

	Throw

	// Assert checks if an expression is true and if it is not, it signals
	// an error/exception, possibly stopping the execution.
	Assert

	// MethodInvocation is the invocation of a method as present in OO
	// languages.
	MethodInvocation
	// MethodInvocationObject is the object where whose method is being
	// invoked.
	MethodInvocationObject
	// MethodInvocationName is the name of the method being invoked.
	MethodInvocationName
	MethodInvocationArgument

	Noop

	// TODO: should we differentiate literal types with specialized literal
	// annotations or combined with other type roles?
	Literal
	NullLiteral
	StringLiteral
	NumberLiteral
	TypeLiteral

	Type
	// TODO: should we distinguist between primitive and builtin types?
	PrimitiveType

	// Assignment represents a variable assignment or binding.
	// The variable that is being assigned to is annotated with the
	// AssignmentVariable role, while the value is annotated with
	// AssignmentValue.
	Assignment
	AssignmentVariable
	AssignmentValue

	// This represents the self-reference of an object instance in
	// one of its methods. This corresponds to the `this` keyword
	// (e.g. Java, C++, PHP), `self` (e.g. Smalltalk, Perl, Swift) and `Me`
	// (e.g. Visual Basic).
	This

	Comment

	// Documentation is a node that represents documentation of another node,
	// such as function or package. Documentation is usually in the form of
	// a string in certain position (e.g. Python docstring) or comment
	// (e.g. Javadoc, godoc).
	Documentation

	// Whitespace
	Whitespace

	// TODO: types
	// TODO: references/pointers
	// TODO: variable declarations
	// TODO: expressions
	// TODO: type parameters

	// TODO: missing mappings from:
	//       Java - try-with-resources
	//       Java - synchronized
	//       Java - class/interface distinction
	//       Go   - goroutines
	//       Go   - defer
	//       Go   - select
	//       Go   - channel operations
)

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
	StartPosition Position
	// EndPosition is the position where this node ends in the original
	// source code file.
	EndPosition Position
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
	err := PreOrderVisit(n, func(path ...*Node) error {
		n := path[len(path)-1]
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

func (n *Node) startPosition() Position {
	if !n.StartPosition.IsEmpty() {
		return n.StartPosition
	}

	var min Position
	for _, c := range n.Children {
		other := c.startPosition()
		if min.IsEmpty() || other.Offset < min.Offset {
			min = other
		}
	}

	return min
}

// String converts the *Node to a string using pretty printing.
func (n *Node) String() string {
	buf := bytes.NewBuffer(nil)
	err := n.Pretty(buf, IncludeAll)
	if err != nil {
		return "error"
	}

	return buf.String()
}

// Pretty writes a pretty string representation of the *Node to a writer.
func (n *Node) Pretty(w io.Writer, includes IncludeFlag) error {
	return printNode(w, 0, n, includes)
}

func printNode(w io.Writer, indent int, n *Node, includes IncludeFlag) error {
	nodeType := n.InternalType
	if !includes.Is(IncludeInternalType) {
		nodeType = "*"
	}

	if _, err := fmt.Fprintf(w, "%s {\n", nodeType); err != nil {
		return err
	}

	istr := strings.Repeat(".  ", indent+1)

	if includes.Is(IncludeAnnotations) && len(n.Roles) > 0 {
		_, err := fmt.Fprintf(w, "%sRoles: %s\n",
			istr,
			rolesToString(n.Roles...),
		)
		if err != nil {
			return err
		}
	}

	if includes.Is(IncludeTokens) && n.Token != nil {
		if _, err := fmt.Fprintf(w, "%sTOKEN \"%s\"\n",
			istr, *n.Token); err != nil {
			return err
		}
	}

	if includes.Is(IncludePositions) && !n.StartPosition.IsEmpty() {
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

	if includes.Is(IncludePositions) && !n.EndPosition.IsEmpty() {
		if _, err := fmt.Fprintf(w, "%EndPosition: {\n", istr); err != nil {
			return err
		}

		if err := printPosition(w, indent+2, n.EndPosition); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
			return err
		}
	}

	if includes.Is(IncludeProperties) && len(n.Properties) > 0 {
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

	if includes.Is(IncludeChildren) && len(n.Children) > 0 {
		if _, err := fmt.Fprintf(w, "%sChildren: {\n", istr); err != nil {
			return err
		}

		if err := printChildren(w, indent+2, n.Children, includes); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
			return err
		}
	}

	if _, err := fmt.Fprintf(w, "%s}\n", istr); err != nil {
		return err
	}

	return nil
}

func printChildren(w io.Writer, indent int, children []*Node, includes IncludeFlag) error {
	istr := strings.Repeat(".  ", indent)

	for idx, child := range children {
		_, err := fmt.Fprintf(w, "%s%d: ",
			istr,
			idx,
		)
		if err != nil {
			return err
		}

		if err := printNode(w, indent, child, includes); err != nil {
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

func printPosition(w io.Writer, indent int, pos Position) error {
	if pos.IsEmpty() {
		return nil
	}

	istr := strings.Repeat(".  ", indent)

	if _, err := fmt.Fprintf(w, "%sOffset: %d\n", istr, pos.Offset); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "%sLine: %d\n", istr, pos.Line); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "%sCol: %d\n", istr, pos.Col)
	return err
}

func rolesToString(roles ...Role) string {
	var strs []string
	for _, r := range roles {
		strs = append(strs, r.String())
	}

	return strings.Join(strs, ",")
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

// Hash returns the hash of the node.
func (n *Node) Hash() Hash {
	return n.HashWith(IncludeChildren)
}

// HashWith returns the hash of the node, computed with the given set of fields.
func (n *Node) HashWith(includes IncludeFlag) Hash {
	//TODO
	return 0
}

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

func (p Position) IsEmpty() bool {
	return p.Offset == 0 && p.Line == 0 && p.Col == 0
}
