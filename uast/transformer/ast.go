package transformer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

func SavePosOffset(vr string) Op {
	return TypedObj(uast.TypePosition, map[string]Op{
		uast.KeyPosOff: Var(vr),
	})
}

func Roles(roles ...role.Role) Op {
	arr := make([]Op, 0, len(roles))
	for _, r := range roles {
		arr = append(arr, Is(uast.Int(r)))
	}
	return Arr(arr...)
}

func ASTMap(name string, native, norm Op) Mapping {
	return Mapping{
		Name: name,
		Steps: []Step{
			{Name: "native", Op: native},
			{Name: "norm", Op: norm},
		},
	}
}
