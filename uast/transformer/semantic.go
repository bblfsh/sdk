package transformer

import (
	"fmt"
	"strings"

	"gopkg.in/bblfsh/sdk.v2/uast"
)

func UASTType(uobj interface{}, op ObjectOp) ObjectOp {
	utyp := uast.TypeOf(uobj)
	if utyp == "" {
		panic(fmt.Errorf("type is not registered: %T", uobj))
	}
	obj := op.Object()
	obj.SetField(uast.KeyType, String(utyp))
	return obj
}

func MapSemantic(name, nativeType string, semType interface{}, pos map[string]string, src, dst ObjectOp) Mapping {
	utyp := uast.TypeOf(semType)
	if strings.HasPrefix(name, " ") {
		name = nativeType + " -> " + utyp + name
	} else if name == "" {
		name = nativeType + " -> " + utyp
	}
	so, do := src.Object(), dst.Object()

	sp := UASTType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	dp := UASTType(uast.Positions{}, Obj{
		uast.KeyStart: Var("start"),
		uast.KeyEnd:   Var("end"),
	}).Object()
	for k, v := range pos {
		sp.SetField(k, Var(v))
		if v != "start" && v != "end" {
			dp.SetField(k, Var(v))
		}
	}
	so.SetField(uast.KeyType, String(nativeType))
	so.SetField(uast.KeyPos, sp)
	do.SetField(uast.KeyPos, dp)
	return Map(name, so, UASTType(semType, do))
}
