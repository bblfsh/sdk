package transformer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

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
	// TopLevelIsRootNode tells ToNode where to find the root node of
	// the AST.  If true, the root will be its input argument. If false,
	// the root will be the value of the only key present in its input
	// argument.
	TopLevelIsRootNode bool
}

func (n ObjectToNode) Do(root uast.Node) (uast.Node, error) {
	root, err := n.transformer().Do(root)
	if err != nil {
		return nil, err
	}
	if obj, ok := root.(uast.Object); ok && !n.TopLevelIsRootNode {
		for _, v := range obj {
			root = v
		}
	}
	return root, nil
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

var _ Transformer = ObjectToNode{}

type RolesDedup struct{}

func (RolesDedup) Do(root uast.Node) (uast.Node, error) {
	nroot, ok := uast.Apply(root, func(n uast.Node) (uast.Node, bool) {
		obj, ok := n.(uast.Object)
		if !ok {
			return n, false
		}
		roles := obj.Roles()
		if len(roles) == 0 {
			return n, false
		}
		m := make(map[role.Role]struct{}, len(roles))
		out := make([]role.Role, 0, len(roles))
		for _, r := range roles {
			if _, ok := m[r]; ok {
				continue
			}
			m[r] = struct{}{}
			out = append(out, r)
		}
		if len(out) == len(roles) {
			return n, false
		}
		nobj := obj.CloneObject()
		nobj.SetRoles(out...)
		return nobj, true
	})
	if ok {
		root = nroot
	}
	return root, nil
}
