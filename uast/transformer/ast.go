package transformer

import (
	"fmt"
	"strings"

	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

// SavePosOffset makes an operation that describes a uast.Position object with Offset field set to a named variable.
func SavePosOffset(vr string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosOff: Var(vr),
	})
}

// SavePosLineCol makes an operation that describes a uast.Position object with Line and Col field set to named variables.
func SavePosLineCol(varLine, varCol string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosLine: Var(varLine),
		uast.KeyPosCol:  Var(varCol),
	})
}

// Roles makes an operation that will check/construct a list of roles.
func Roles(roles ...role.Role) ArrayOp {
	arr := make([]Op, 0, len(roles))
	for _, r := range roles {
		arr = append(arr, Is(uast.String(r.String())))
	}
	return Arr(arr...)
}

// AppendRoles can be used to append more roles to an output of a specific operation.
func AppendRoles(old ArrayOp, roles ...role.Role) ArrayOp {
	if len(roles) == 0 {
		return old
	}
	return AppendArr(old, Roles(roles...))
}

// ASTMap is a helper for creating a two-way mapping between AST and its normalized form.
func ASTMap(name string, native, norm Op) Mapping {
	return Mapping{
		Name: name,
		Steps: []Step{
			{Name: "native", Op: native},
			{Name: "norm", Op: norm},
		},
	}
}

// RolesField will create a roles field that appends provided roles to existing ones.
// In case no roles are provided, it will save existing roles, if any.
func RolesField(vr string, roles ...role.Role) Field {
	return RolesFieldOp(vr, nil, roles...)
}

// RolesFieldOp is like RolesField but allows to specify custom roles op to use.
func RolesFieldOp(vr string, op ArrayOp, roles ...role.Role) Field {
	if len(roles) == 0 && op == nil {
		return Field{
			Name:     uast.KeyRoles,
			Op:       Var(vr),
			Optional: vr + "_exists",
		}
	}
	var rop ArrayOp
	if len(roles) != 0 && op != nil {
		rop = AppendRoles(op, roles...)
	} else if op != nil {
		rop = op
	} else {
		rop = Roles(roles...)
	}
	return Field{
		Name: uast.KeyRoles,
		Op: If(vr+"_exists",
			Append(NotEmpty(Var(vr)), rop),
			rop,
		),
	}
}

// ASTObjectLeft construct a native AST shape for a given type name.
func ASTObjectLeft(typ string, ast ObjectOp) ObjectOp {
	a := ast.Object()
	if _, ok := a.GetField(uast.KeyRoles); ok {
		panic("unexpected roles filed")
	}
	if typ != "" {
		a.SetField(uast.KeyType, String(typ))
	}
	a.SetFieldObj(RolesField(typ + "_roles"))
	return Part("_", a)
}

// ASTObjectRight constructs an annotated native AST shape with specific roles.
func ASTObjectRight(typ string, norm ObjectOp, rop ArrayOp, roles ...role.Role) ObjectOp {
	return ASTObjectRightCustom(typ, norm, nil, rop, roles...)
}

// RolesByType is a function for getting roles for specific AST node type.
type RolesByType func(typ string) role.Roles

// ASTObjectRightCustom is like ASTObjectRight but allow to specify additional roles for each type.
func ASTObjectRightCustom(typ string, norm ObjectOp, fnc RolesByType, rop ArrayOp, roles ...role.Role) ObjectOp {
	b := norm.Object()
	if _, ok := b.GetField(uast.KeyRoles); ok {
		panic("unexpected roles filed")
	}
	if typ != "" {
		b.SetField(uast.KeyType, String(typ)) // TODO: "<lang>:" namespace
	}
	// it merges 3 slices:
	// 1) roles saved from left side (if any)
	// 2) static roles from arguments
	// 3) roles from conditional operation
	if fnc != nil {
		if static := fnc(typ); len(static) > 0 {
			roles = append([]role.Role{}, roles...)
			roles = append(roles, static...)
		}
	}
	b.SetFieldObj(RolesFieldOp(typ+"_roles", rop, roles...))
	return Part("_", b)
}

// ObjectRoles creates a shape that adds additional roles to an object.
// Should only be used in other object fields, since it does not set any type constraints.
func ObjectRoles(vr string, roles ...role.Role) Op {
	return ObjectRolesCustom(vr, nil, roles...)
}

// ObjectRolesCustom is like ObjectRoles but allows to apecify additional conatraints for object.
func ObjectRolesCustom(vr string, obj ObjectOp, roles ...role.Role) Op {
	return ObjectRolesCustomOp(vr, obj, nil, roles...)
}

// ObjectRolesCustomOp is like ObjectRolesCustom but allows to apecify a custom roles lookup.
func ObjectRolesCustomOp(vr string, obj ObjectOp, rop ArrayOp, roles ...role.Role) Op {
	f := RolesFieldOp(vr+"_roles", rop, roles...)
	if obj == nil {
		obj = Fields{f}
	} else {
		o := obj.Object()
		o.SetFieldObj(f)
		obj = o
	}
	return Part(vr, obj)
}

// EachObjectRoles is a helper that constructs Each(ObjectRoles(roles)) operation.
// It will annotate all elements of the array with a specified list of roles.
func EachObjectRoles(vr string, roles ...role.Role) Op {
	return eachObjectRolesByType(vr, nil, roles...)
}

// EachObjectRolesByType is like EachObjectRoles but adds additional roles for each type specified in the map.
// EachObjectRolesByType should always be paired on both side of transform because it uses variables to store type info.
func EachObjectRolesByType(vr string, types map[string][]role.Role, roles ...role.Role) Op {
	if types == nil {
		types = make(map[string][]role.Role)
	}
	return eachObjectRolesByType(vr, types, roles...)
}

func eachObjectRolesByType(vr string, types map[string][]role.Role, roles ...role.Role) Op {
	var (
		obj ObjectOp
		rop ArrayOp
	)
	if types != nil {
		tvar := vr + "_type"
		obj = Obj{
			uast.KeyType: Var(tvar),
		}
		if len(types) != 0 {
			cases := make(map[uast.Value]ArrayOp, len(types))
			for typ, arr := range types {
				var key uast.Value
				if typ != "" {
					key = uast.String(typ)
				}
				cases[key] = Roles(arr...)
			}
			rop = LookupArrOpVar(tvar, cases)
		}
	}
	return Each(vr+"_arr", ObjectRolesCustomOp(vr, obj, rop, roles...))
}

// OptObjectRoles is like ObjectRoles, but marks an object as optional.
func OptObjectRoles(vr string, roles ...role.Role) Op {
	return Opt(vr+"_set", ObjectRoles(vr, roles...))
}

// Operator is a helper to make an AST node describing an operator.
func Operator(vr string, lookup map[uast.Value]ArrayOp, roles ...role.Role) ObjectOp {
	roles = append([]role.Role{
		role.Expression, role.Operator,
	}, roles...)
	var opRoles Op = Roles(roles...)
	if lookup != nil {
		opRoles = AppendRoles(
			LookupArrOpVar(vr, lookup),
			roles...,
		)
	}
	return Fields{
		{Name: uast.KeyType, Op: String(uast.TypeOperator)},
		{Name: uast.KeyToken, Op: Var(vr)},
		{Name: uast.KeyRoles, Op: opRoles},
	}
}

func uncomment(s string) (string, error) {
	// Remove // and /*...*/ from comment nodes' tokens
	if strings.HasPrefix(s, "//") {
		s = s[2:]
	} else if strings.HasPrefix(s, "/*") {
		s = s[2 : len(s)-2]
	}
	return s, nil
}

func comment(s string) (string, error) {
	if strings.Contains(s, "\n") {
		return "/*" + s + "*/", nil
	}
	return "//" + s, nil
}

// UncommentCLike removes // and /* */ symbols from a given string variable.
func UncommentCLike(vr string) Op {
	return StringConv(Var(vr), uncomment, comment)
}

// MapAST is a helper for describing a single AST transformation for a given node type.
func MapAST(typ string, ast, norm ObjectOp, roles ...role.Role) Mapping {
	return MapASTCustom(typ, ast, norm, nil, roles...)
}

// MapASTCustom is like MapAST, but allows to specify additional operation for adding roles.
func MapASTCustom(typ string, ast, norm ObjectOp, rop ArrayOp, roles ...role.Role) Mapping {
	return ASTMap(typ,
		ASTObjectLeft(typ, ast),
		ASTObjectRight(typ, norm, rop, roles...),
	)
}

// MapASTCustomType is like MapASTCustom, but allows to specify additional roles for each type.
func MapASTCustomType(typ string, ast, norm ObjectOp, fnc RolesByType, rop ArrayOp, roles ...role.Role) Mapping {
	return ASTMap(typ,
		ASTObjectLeft(typ, ast),
		ASTObjectRightCustom(typ, norm, fnc, rop, roles...),
	)
}

// ObjRoles is a helper type that stores a mapping from field names to their roles.
type ObjRoles map[string][]role.Role

var _ ASTMapFunc = MapASTCustom

// ASTMapFunc is a signature for functions that maps two AST shapes for a specific type and can append roles to it.
type ASTMapFunc func(typ string, ast, norm ObjectOp, rop ArrayOp, roles ...role.Role) Mapping

// AnnotateTypeCustom is like AnnotateType but allows to specify custom roles operation as well as a mapper function.
func AnnotateTypeCustom(mapAST ASTMapFunc, typ string, fields ObjRoles, rop ArrayOp, roles ...role.Role) Mapping {
	m := make(FieldRoles, len(fields))
	for name, roles := range fields {
		m[name] = FieldRole{Opt: true, Roles: roles}
	}
	return AnnotateTypeFieldsCustom(mapAST, typ, m, rop, roles...)
}

// AnnotateType is a helper to assign roles to specific fields. All fields are assumed to be optional and should be objects.
func AnnotateType(typ string, fields ObjRoles, roles ...role.Role) Mapping {
	return AnnotateTypeCustom(nil, typ, fields, nil, roles...)
}

// FieldRole is a list of operations that can be applied to an object field.
type FieldRole struct {
	Rename string // rename the field to this name in resulting tree

	Skip bool // omit this field in the resulting tree
	Add  bool // create this field in the resulting tree

	Opt   bool        // field can be nil
	Arr   bool        // field is an array; apply roles or custom operation to each element
	Op    Op          // use this operation for the field on both sides of transformation
	Roles []role.Role // list of roles to append to the field; has no effect if Op is set
}

func (f FieldRole) validate() error {
	if f.Arr && f.Opt {
		return fmt.Errorf("field should either be a list or optional")
	}
	if len(f.Roles) == 0 && f.Op == nil && (f.Opt || f.Arr) {
		return fmt.Errorf("either roles or operation should be set to use Opt or Arr")
	}
	if f.Skip && (f.Opt || f.Arr || len(f.Roles) != 0) {
		return fmt.Errorf("skip cannot be used with other operations")
	}
	if f.Skip && (f.Rename != "" && !f.Add) {
		return fmt.Errorf("rename can only be used with skip when Add is set")
	}
	return nil
}

func (f FieldRole) build(name string) (names [2]string, ops [2]Op, _ error) {
	if err := f.validate(); err != nil {
		return names, ops, err
	}
	rname := name
	if f.Rename != "" {
		rname = f.Rename
	}
	vr := name + "_var"
	var l, r Op
	if f.Op != nil {
		l, r = f.Op, f.Op
	} else if len(f.Roles) == 0 {
		l, r = Var(vr), Var(vr)
	} else {
		var fnc func(vr string, roles ...role.Role) Op
		if f.Arr {
			fnc = EachObjectRoles
		} else if f.Opt {
			fnc = OptObjectRoles
		} else {
			fnc = ObjectRoles
		}
		l, r = fnc(vr), fnc(vr, f.Roles...)
	}
	if f.Skip {
		l = AnyVal(nil)
	}
	if f.Skip || !f.Add {
		names[0] = name
		ops[0] = l
	}
	if !f.Skip || f.Add {
		names[1] = rname
		ops[1] = r
	}
	return names, ops, nil
}

// FieldRoles is a helper type that stores a mapping from field names to operations that needs to be applied to it.
type FieldRoles map[string]FieldRole

// AnnotateTypeFieldsCustom is like AnnotateTypeFields but allows to specify custom roles operation as well as a mapper function.
func AnnotateTypeFieldsCustom(mapAST ASTMapFunc, typ string, fields FieldRoles, rop ArrayOp, roles ...role.Role) Mapping {
	if mapAST == nil {
		mapAST = MapASTCustom
	}
	left := make(Obj, len(fields))
	right := make(Obj, len(fields))
	for name, f := range fields {
		names, ops, err := f.build(name)
		if err != nil {
			panic(fmt.Errorf("field %q: %v", name, err))
		}
		if names[0] != "" {
			left[names[0]] = ops[0]
		}
		if names[1] != "" {
			right[names[1]] = ops[1]
		}
	}
	return mapAST(typ, left, right, rop, roles...)
}

// AnnotateTypeFields is a more advanced version of AnnotateType and allows to apply more complex operations to fields.
// See FieldRole for a list of supported operations.
func AnnotateTypeFields(typ string, fields FieldRoles, roles ...role.Role) Mapping {
	return AnnotateTypeFieldsCustom(nil, typ, fields, nil, roles...)
}

// StringToRolesMap is a helper to generate an array operation map that can be used for Lookup
// from a map from string values to roles.
func StringToRolesMap(m map[string][]role.Role) map[uast.Value]ArrayOp {
	out := make(map[uast.Value]ArrayOp, len(m))
	for tok, roles := range m {
		out[uast.String(tok)] = Roles(roles...)
	}
	return out
}
