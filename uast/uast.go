package uast

//proteus:generate
type Role int8

const (
	PackageDeclaration Role = iota
	FunctionDeclaration
)

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
