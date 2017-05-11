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
type Role int16

const (
	_ = iota

	// SimpleIdentifier is the most basic form of identifier, used for variable
	// names, functions, packages, etc.
	SimpleIdentifier Role = iota
	// QualifiedIdentifier is form of identifier composed of multiple
	// SimpleIdentifier. One main identifier (usually the last one) and one
	// or more qualifiers.
	QualifiedIdentifier

	// BinaryExpression is the parent node of all binary expressions of any type. It must have
	// BinaryExpressionLeft, BinaryExpressionRight and BinaryExpressionOp children.
	// Those children must have aditional roles specifying the specific type (e.g. Expression,
	// QualifiedIdentifier or Literal for the left and right nodes and one of the specific operator roles
	// for the binary operator). BinaryExpresion can be considered a derivation of Expression and thus
	// could be its child or implemented as an additional node.
	BinaryExpression
	BinaryExpressionLeft
	BinaryExpressionRight
	BinaryExpressionOp

	// Infix should mark the nodes which are parents of expression nodes using infix notation, e.g.: a+b.
	// Nodes without Infix or Postfix mark are considered in prefix order by default.
	Infix
	// Postfix should mark the nodes which are parents of nodes using postfix notation, e.g.: ab+.
	// Nodes without Infix or Postfix mark are considered in prefix order by default.
	Postfix

	// Binary bitwise operators, used to alterate bits on numeral variables

	// OpBitwiseLeftShift is the binary bitwise shift to the left operator (i.e. << in most languages)
	OpBitwiseLeftShift
	// OpBitwiseRightShift is the binary bitwise shift to the right operator (i.e. >> in most languages)
	OpBitwiseRightShift
	// OpBitwiseUnsignedRightShift is the binary bitwise unsigned shift to the
	// right operator (e.g. >>> in Java or C#)
	OpBitwiseUnsignedRightShift
	// OpBitwiseOr is the binary bitwise OR operator  (i.e. | in most languages)
	OpBitwiseOr
	// OpBitwiseXor is the binary bitwise Xor operator  (i.e. ~ in most languages)
	OpBitwiseXor
	// OpBitwiseAnd is the binary bitwise And/complement operator  (i.e. & in most languages)
	OpBitwiseAnd

	Expression
	Statement

	// Comparison operators. Usually inside one of the Condition nodes, but could also be in
	// an expression of boolean value by itself. The tested expressions will be its siblings.

	// OpEqual is the operator that tests for logical equality between two expressions
	OpEqual
	// OpEqual is the operator that tests for logical inequality between two expressions
	// (i.e. != or != or <> in most languages).
	OpNotEqual
	// OpEqual is the operator that tests if the expression on the left is worth logically less
	// than the expression on the right. (i.e. < in most languages).
	OpLessThan
	// OpEqual is the operator that tests if the expression on the left is worth logically less
	// or has equality with the expression on the right. (i.e. >= in most languages).
	OpLessThanEqual
	// OpEqual is the operator that tests if the expression on the left is worth logically more
	// than the expression on the right. (i.e. > in most languages).
	OpGreaterThan
	// OpEqual is the operator that tests if the expression on the left is worth logically more
	// or has equality with the expression on the right. (i.e. >= in most languages).
	OpGreaterThanEqual
	// OpSame tests if the result of the expressions tested is the same object, like the "is"
	// operator in node or === in Javascript.
	OpSame
	// OpNotSame tests if the result of the expressions tested are different objects, like the "is not"
	// operator in node or !== in Javascript.
	OpNotSame
	// OpContains tests if the left expression result is contained inside, or has an item contained
	// with equality, the result of the expression of the right which usually will be a container type
	// (e.g. "in" in Python).
	OpContains
	// OpNotContains tests if the left expression result is not contained inside
	// the result of the expression of the right which usually will be a container type
	// (e.g. "not in" in Python).
	OpNotContains

	// Unary operators. These will have a single child node that will apply the operator over it.
	// TODO: use BooleanNot to implement the UnaryNot?

	// OpPreIncrement increments in place the value before it is evaluated. It's
	// typical of C-inspired languages (e. g. ++x).
	OpPreIncrement
	// OpPostIncrement increments in place the value after it is evaluated. It's
	// typical of C-inspired languages (e. g. x++).
	OpPostIncrement
	// OpPreDecrement decrement in place the value before it is evaluated. It's
	// typical of C-inspired languages (e. g. --x).
	OpPreDecrement
	// OpPostDecrement decrement in place the value after it is evaluated. It's
	// typical of C-inspired languages (e. g. x--).
	OpPostDecrement
	// OpNegative changes the sign of the numeric type (e. g. -x in most languages).
	OpNegative
	// OpPositive usually is a no-op for basic numeric types but exists in the AST of some languages.
	// On some languages like C it could perform an aritmetic conversion to a signed type without
	// changing the sign or could be overloaded (e. g. +x).
	OpPositive
	// OpBitwiseComplement will invert all the bits of a type. (e. g. ~x in C-inspired languages).
	OpBitwiseComplement
	// OpDereference will get the actual value pointed by a pointer or reference type (e.g. *x).
	OpDereference
	// OpTakeAddress will get the memory address of the associated variable which will usually be
	// stored in a pointer or reference type (e. g. &x).
	OpTakeAddress

	// File is the root node of a single file AST.
	File

	// Binary boolean operators, like

	// OpBooleanAnd is the boolean AND operator (i.e. "and" or && in most languages)
	OpBooleanAnd
	// OpBooleanOr is the boolean OR operator (i.e. "or" or || in most languages)
	OpBooleanOr
	// OpBooleanNot is the boolean NOT operator (i.e. "NOT" or ! in most languages)
	OpBooleanNot
	// OpBooleanXor is the boolean XOR operator (i.e. "XOR" or ^ in most languages)
	OpBooleanXor

	// Binary aritmethic operators. Examples with C operators.
	// TODO: should we have division and FloorDivision like Python or Nim?
	// TODO: should we had the pow operator that some languages have?

	// OpAdd is the binary add operator (i.e. + in most languages).
	OpAdd
	// OpSubstract is the binary subtract operator (i.e. - in most languages).
	OpSubstract
	// OpMultiply is the binary multiply operator (i.e. * in most languages).
	OpMultiply
	// OpDivide is the binary division operator (i.e. / in most languages).
	OpDivide
	// OpMod is the binary division module operator (i.e. % or "mod" in most languages).
	OpMod

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

	// TODO: argument type declarations, return value, body, etc.

	// FunctionDeclaration is the parent node of all function or method declarations. It should have a
	// FunctionDeclarationName, a FunctionDeclarationBody (except for pure declarations like the ones in C/C++
	// header files or forward declarations in other languages) and, if the function has formal arguments,
	// FunctionDeclarationArgument children.
	FunctionDeclaration
	// FunctionDeclarationBody is the grouping node for all nodes in the function body.
	FunctionDeclarationBody
	// FunctionDeclarationName contains the unqualified name of the function.
	FunctionDeclarationName
	// FunctionDeclarationReceiver is the target Type of a method or struct.
	FunctionDeclarationReceiver
	// FunctionDeclarationArgument is the parent node for the function formal arguments. The name will be
	// specified as the token of the child FunctionDeclarationArgumentName and depending on the language it
	// could have one or more child nodes of different types to implement them in the UAST like
	// FunctionDeclarationArgumentDefaultValue, type declarations (TODO), annotations (TODO), etc.
	//FunctionDeclarationArguments
	FunctionDeclarationArgument
	// FunctionDeclarationArgumentName is the symbolic name of the argument. On languages that support
	// argument passing by name this will be the name used by the CallNamedArgument roles.
	FunctionDeclarationArgumentName
	// For languages that support setting a default value for a formal argument,
	// FunctionDeclarationArgumentDefaultValue is the node that contains the default value.
	// Depending on the language his child node representing the actual value could be some kind or
	// literal or even expressions that can resolved at runtime (if interpreted) or compile time.
	FunctionDeclarationArgumentDefaultValue
	// FunctionDeclarationVarArgsList is the node representing whatever syntax the language has to
	// indicate that from that point in the argument list the function can get a variable number
	// of arguments (e.g. "..." in C-ish languages, "Object..." in Java, "*args" in Python, etc).
	FunctionDeclarationVarArgsList

	// TypeDeclaration is the declaration of a type. It could be a class or
	// interface in Java, a struct, interface or alias in Go, etc. Except for pure or forward declarations
	// it will usually have a TypeDeclarationBody child and for OOP languages a TypeDeclarationBases and/or
	// TypeDeclarationInterfaces.
	TypeDeclaration
	TypeDeclarationBody
	// TypeDeclarationBases are the Types that the current inherits from in OOP languages.
	TypeDeclarationBases
	// TypeDeclarationImplements are the Types (usually interfaces) that the Type implements.
	TypeDeclarationImplements

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
	Goto

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

	// BooleanLiteral is a boolean literal. It is expected that BooleanLiteral
	// nodes contain a token with some form of boolean literal (e.g. true,
	// false, yes, no, 1, 0).
	BooleanLiteral
	// ByteLiteral is a single-byte literal. For example, in Rust.
	ByteLiteral
	// ByteStringLiteral is a literal for a raw byte string. For example, in Rust.
	ByteStringLiteral
	// CharacterLiteral is a character literal. It is expected that
	// CharacterLiteral nodes contain a token with a single character with
	// optional quoting (e.g. c, 'c', "c").
	CharacterLiteral
	// ListLiteral is a literal array or list.
	ListLiteral
	// MapLiteral is a literal map-like structure.
	MapLiteral
	// NullLiteral is a null literal. It is expected that NullLiteral nodes
	// contain a token equivalent to null (e.g. null, nil, None).
	NullLiteral
	// NumberLiteral is a numeric literal. This applies to any numeric literal
	// whether it is integer or float, any base, scientific notation or not,
	// etc.
	NumberLiteral
	// RegexpLiteral is a literal for a regular expression.
	RegexpLiteral
	// SetLiteral is a literal for a set. For example, in Python 3.
	SetLiteral
	// StringLiteral is a string literal. This applies both to single-line and
	// multi-line literals and it does not imply any particular encoding.
	//
	// TODO: Decide what to do with interpolated strings.
	StringLiteral
	// TupleLiteral is a literal for a tuple. For example, in Python and Scala.
	TupleLiteral
	// TypeLiteral is a literal that identifies a type. It might contain a
	// token with the type literal itself, or children that define the type.
	TypeLiteral
	// OtherLiteral is a literal of a type not covered by other literal
	// annotations.
	OtherLiteral

	// MapEntry is the expression pairing a map key and a value usually on MapLiteral expressions. It must
	// have both a MapKey and a MapValue children (e.g. {"key": "value", "otherkey": "otherval"} in Python).
	MapEntry
	MapKey
	MapValue

	Type
	// TODO: should we distinguish between primitive and builtin types?
	PrimitiveType

	// Assignment represents a variable assignment or binding.
	// The variable that is being assigned to is annotated with the
	// AssignmentVariable role, while the value is annotated with
	// AssignmentValue.
	Assignment
	AssignmentVariable
	AssignmentValue

	// AugmentedAssignment is an augmented assignment usually combining the equal operator with
	// another one (e. g. +=, -=, *=, etc). It is expected that children contains an
	// AugmentedAssignmentOperator with a child or aditional role for the specific Bitwise or
	// Arithmetic operator used. The AugmentedAssignmentVariable and AugmentedAssignmentValue roles
	// have the same meaning than in Assignment.
	AugmentedAssignment
	AugmentedAssignmentOperator
	AugmentedAssignmentVariable
	AugmentedAssignmentValue

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
	// TODO: references/pointer member access
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
	StartPosition *Position `json:",omitempty"`
	// EndPosition is the position where this node ends in the original
	// source code file.
	EndPosition *Position `json:",omitempty"`
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

const (
	// InternalRoleKey is a key string uses in properties to use the internal
	// role of a node in the AST, if any.
	InternalRoleKey = "internalRole"
)

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
