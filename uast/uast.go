// Package uast defines a UAST (Universal Abstract Syntax Tree) representation
// and operations to manipulate them.
package uast

import (
	"bytes"
)

// go:generate proteus proto -p github.com/bblfsh/sdk/uast -f $GOPATH/src/github.com/bblfsh/sdk/protos

// Hash is a hash value.
type Hash uint32

// Role is the main UAST annotation. It indicates that a node in an AST can
// be interpreted as acting with certain language-independent role.
//
//proteus:generate
//go:generate stringer -type=Role
type Role int8

const (
	_ = iota

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

	// VisibleFromInstance marks modifiers that declare visibility from instance.
	VisibleFromInstance
	// VisibleFromType marks modifiers that declare visibility from the same
	// type (e.g. class, trait).
	// Implies VisibleFromInstance.
	VisibleFromType
	// VisibleFromSubtype marks modifiers that declare visibility from
	// subtypes (e.g. subclasses).
	// Implies VisibleFromInstance and VisibleFromType.
	VisibleFromSubtype
	// VisibleFromSubpackage marks modifiers that declare visibility from the
	// same package.
	VisibleFromPackage
	// VisibleFromSubpackage marks modifiers that declare visibility from
	// subpackages.
	// Implies VisibleFromInstance, VisibleFromType and VisibleFromPackage.
	VisibleFromSubpackage
	// VisibleFromModule marks modifiers that declare visibility from the
	// same module (e.g. Java JAR).
	// Implies VisibleFromInstance and VisibleFromType.
	VisibleFromModule
	// VisibleFromFriend marks modifiers that declare visibility from friends
	// (e.g. C++ friends).
	// Implies VisibleFromInstance and VisibleFromType.
	VisibleFromFriend
	// VisibleFromWorld implies full public visibility. Implies all other
	// visibility levels.
	VisibleFromWorld

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

	// Call is any call, whether it is a function, procedure, method or macro.
	// In its simplest form, a call will have a single child with a function
	// name (CallCallee). Arguments are marked with CallPositionalArgument
	// and CallNamedArgument. In OO languages there is usually a CallReceiver
	// too.
	Call
	// CallReceiver is an optional expression receiving the call. This
	// corresponds to the method invocant in OO languages, receiving in Go, etc.
	CallReceiver
	// CallCallee is the callable being called. It might be the name of a
	// function or procedure, it might be a method, it might a simple name
	// or qualified with a namespace.
	CallCallee
	// CallPositionalArgument is a positional argument in a call.
	CallPositionalArgument
	// CallNamedArgument is a named argument in a call. It should have a
	// child with role CallNamedArgumentName and another child with role
	// CallNamedArgumentValue.
	CallNamedArgument
	// CallNamedArgumentName is the name of a named argument.
	CallNamedArgumentName
	// CallNamedArgumentValue is the value of a named argument.
	CallNamedArgumentValue

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

// IsEmpty returns true if the position is empty.
func (p Position) IsEmpty() bool {
	return p.Offset == 0 && p.Line == 0 && p.Col == 0
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
	StartPosition Position `json:",omitempty"`
	// EndPosition is the position where this node ends in the original
	// source code file.
	EndPosition Position `json:",omitempty"`
	// Roles is a list of Role that this node has. It is a language-independent
	// annotation.
	Roles []Role `json:",omitempty"`
}

// NewNode creates a new empty *Node.
func NewNode() *Node {
	return &Node{
		Properties: make(map[string]string, 0),
	}
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

// String converts the *Node to a string using pretty printing.
func (n *Node) String() string {
	buf := bytes.NewBuffer(nil)
	err := Pretty(n, buf, IncludeAll)
	if err != nil {
		return "error"
	}

	return buf.String()
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
	dst := make(Path, len(p) + 1)
	copy(dst, p)
	dst[len(p)] = n
	return dst
}

// Parent returns the path of the parent of this node.
func (p Path) Parent() Path {
	if len(p) <= 1 {
		return Path(nil)
	}

	dst := make(Path, 0, len(p) - 1)
	copy(dst, p)
	return dst
}
