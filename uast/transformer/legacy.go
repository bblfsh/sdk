package transformer

import (
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/role"
)

const dedupCloneObj = false

var _ Transformer = ResponseMetadata{}

// ResponseMetadata is a transformation that is applied to the root of AST tree to trim any metadata that might be there.
type ResponseMetadata struct {
	// TopLevelIsRootNode tells ToNode where to find the root node of
	// the AST.  If true, the root will be its input argument. If false,
	// the root will be the value of the only key present in its input
	// argument.
	TopLevelIsRootNode bool
}

// Do applies the transformation described by this object.
func (n ResponseMetadata) Do(root nodes.Node) (nodes.Node, error) {
	if obj, ok := root.(nodes.Object); ok && !n.TopLevelIsRootNode && len(obj) == 1 {
		for _, v := range obj {
			root = v
			break
		}
	}
	return root, nil
}

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
	// LineKey is the key used in the native AST to indicate
	// the line number where the code mapped to the AST node starts.
	LineKey string
	// EndLineKey is the key used in the native AST to indicate
	// the line number where the code mapped to the AST node ends.
	EndLineKey string
	// ColumnKey is a key that indicates the column inside the line
	ColumnKey string
	// EndColumnKey is a key that indicates the column inside the line where the node ends.
	EndColumnKey string
}

// Mapping construct a transformation from ObjectToNode definition.
func (n ObjectToNode) Mapping() Mapping {
	var (
		ast = make(Obj)
		// ->
		norm, normPos = make(Obj), make(Obj)
	)

	if n.InternalTypeKey != "" {
		const vr = "itype"
		ast[n.InternalTypeKey] = Var(vr)
		norm[uast.KeyType] = Var(vr)
	}

	if n.OffsetKey != "" {
		const vr = "pos_off_start"
		ast[n.OffsetKey] = Var(vr)
		normPos[uast.KeyStart] = SavePosOffset(vr)
	}
	if n.EndOffsetKey != "" {
		const vr = "pos_off_end"
		ast[n.EndOffsetKey] = Var(vr)
		normPos[uast.KeyEnd] = SavePosOffset(vr)
	}
	if n.LineKey != "" && n.ColumnKey != "" {
		const (
			vrl = "pos_line_start"
			vrc = "pos_col_start"
		)
		ast[n.LineKey] = Var(vrl)
		ast[n.ColumnKey] = Var(vrc)
		normPos[uast.KeyStart] = SavePosLineCol(vrl, vrc)
	} else if (n.LineKey != "" && n.ColumnKey == "") || (n.LineKey == "" && n.ColumnKey != "") {
		panic("both LineKey and ColumnKey should either be set or not")
	}
	if n.EndLineKey != "" && n.EndColumnKey != "" {
		const (
			vrl = "pos_line_end"
			vrc = "pos_col_end"
		)
		ast[n.EndLineKey] = Var(vrl)
		ast[n.EndColumnKey] = Var(vrc)
		normPos[uast.KeyEnd] = SavePosLineCol(vrl, vrc)
	} else if (n.EndLineKey != "" && n.EndColumnKey == "") || (n.EndLineKey == "" && n.EndColumnKey != "") {
		panic("both EndLineKey and EndColumnKey should either be set or not")
	}
	if len(normPos) != 0 {
		norm[uast.KeyPos] = UASTType(uast.Positions{}, normPos)
	}
	return MapPart("other", MapObj(ast, norm))
}

// RolesDedup is an irreversible transformation that removes duplicate roles from AST nodes.
func RolesDedup() TransformFunc {
	return TransformFunc(func(n nodes.Node) (nodes.Node, bool, error) {
		obj, ok := n.(nodes.Object)
		if !ok {
			return n, false, nil
		}
		roles := uast.RolesOf(obj)
		if len(roles) == 0 {
			return n, false, nil
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
			return n, false, nil
		}
		if dedupCloneObj {
			obj = obj.CloneObject()
		}
		obj[uast.KeyRoles] = uast.RoleList(out...)
		return obj, dedupCloneObj, nil
	})
}
