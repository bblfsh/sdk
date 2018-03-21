package transformer

import "gopkg.in/bblfsh/sdk.v1/uast"

var _ Transformer = ObjectToNode{}

// ObjectToNode transform trees that are represented as nested JSON objects.
// That is, an interface{} containing maps, slices, strings and integers. It
// then converts from that structure to *Node.
type ObjectToNode struct {
	// InternalTypeKey is the name of the key that the native AST uses
	// to differentiate the type of the AST nodes. This internal key will then be
	// checkable in the AnnotationRules with the `HasInternalType` predicate. This
	// field is mandatory.
	InternalTypeKey string
	// OffsetKey is the key used in the native AST to indicate the absolute offset,
	// from the file start position, where the code mapped to the AST node starts.
	OffsetKey string
	// EndOffsetKey is the key used in the native AST to indicate the absolute offset,
	// from the file start position, where the code mapped to the AST node ends.
	EndOffsetKey string
}

func (n ObjectToNode) Do(root uast.Node) (uast.Node, error) {
	return n.transformer().Do(root)
}
func (n ObjectToNode) transformer() Transformer {
	ast := make(Obj)
	// ->
	norm := make(Obj)

	if n.InternalTypeKey != "" {
		const vr = "itype"
		ast[n.InternalTypeKey] = Var(vr)
		norm[uast.KeyType] = Var(vr)
	}
	if n.OffsetKey != "" {
		const vr = "pos_off_start"
		ast[n.OffsetKey] = Var(vr)
		norm[uast.KeyStart] = SavePosOffset(vr)
	}
	if n.EndOffsetKey != "" {
		const vr = "pos_off_end"
		ast[n.EndOffsetKey] = Var(vr)
		norm[uast.KeyEnd] = SavePosOffset(vr)
	}
	return ASTMap("ObjectToNode",
		Part("other", ast),
		Part("other", norm),
	)
}
