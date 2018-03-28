package transformer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

// SavePosOffset makes an operation that describes a uast.Position object with Offset field set to a named variable.
func SavePosOffset(vr string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosOff: Var(vr),
	})
}

// SavePosLine makes an operation that describes a uast.Position object with Line field set to a named variable.
func SavePosLine(vr string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosLine: Var(vr),
	})
}

// SavePosCol makes an operation that describes a uast.Position object with Col field set to a named variable.
func SavePosCol(vr string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosCol: Var(vr),
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
func AppendRoles(old Op, roles ...role.Role) Op {
	if len(roles) == 0 {
		return old
	}
	return Append(old, Roles(roles...))
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
	if len(roles) == 0 {
		return Field{
			Name:     uast.KeyRoles,
			Op:       Var(vr),
			Optional: vr + "_exists",
		}
	}
	return Field{
		Name: uast.KeyRoles,
		Op: If(vr+"_exists",
			AppendRoles(NotEmpty(Var(vr)), roles...),
			Roles(roles...),
		),
	}
}
