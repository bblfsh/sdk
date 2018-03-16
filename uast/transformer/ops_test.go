package transformer

import (
	"testing"

	"github.com/stretchr/testify/require"
	u "gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/src-d/go-errors.v1"
)

func arrObjInt(key string, v int) func() u.Node {
	return arrObjVal(key, u.Int(v))
}

func arrObjStr(key string, v string) func() u.Node {
	return arrObjVal(key, u.String(v))
}

func arrObjVal(key string, v u.Value) func() u.Node {
	return func() u.Node {
		return u.List{
			u.Object{key: v},
		}
	}
}

func arrObjVal2(key1, key2 string, v1, v2 u.Value) func() u.Node {
	return func() u.Node {
		return u.List{
			u.Object{key1: v1, key2: v2},
		}
	}
}

var opCases = []struct {
	name     string
	inp, exp func() u.Node
	src, dst Op
	err      *errors.Kind
	noRev    bool // should only be set in exceptional cases
}{
	{
		name: "is",
		inp:  arrObjInt("v", 1),
		src:  Is(u.Int(1)),
		dst:  Is(u.Int(2)),
		exp:  arrObjInt("v", 2),
	},
	{
		name: "var all",
		inp:  arrObjInt("v", 1),
		src:  Var("n"),
		dst:  Var("n"),
	},
	{
		name: "obj has",
		inp:  arrObjInt("v", 1),
		src:  Obj(Has("v", u.Int(1))),
		dst:  Obj(Has("v2", u.Int(2))),
		exp:  arrObjInt("v2", 2),
	},
	{
		name: "has nil",
		inp:  arrObjVal("v", nil),
		src:  Obj(Has("v", nil)),
		dst:  Obj(Has("v2", u.Int(2))),
		exp:  arrObjInt("v2", 2),
	},
	{
		name: "obj save",
		inp:  arrObjInt("v", 1),
		src:  Obj(Save("v", "x")),
		dst:  Obj(Save("v2", "x")),
		exp:  arrObjInt("v2", 1),
	},
	{
		name: "save nil",
		inp:  arrObjVal("v", nil),
		src:  Obj(Save("v", "x")),
		dst:  Obj(Save("v2", "x")),
		exp:  arrObjVal("v2", nil),
	},
	{
		name: "arr save",
		inp:  arrObjInt("v", 1),
		src:  One(Obj(Save("v", "x"))),
		dst:  One(Obj(Save("v2", "x"))),
		exp:  arrObjInt("v2", 1),
	},
	{
		name: "lookup save",
		inp:  arrObjInt("v", 1),
		src: Obj(
			Out("v", LookupVar("x", map[u.Value]u.Value{
				u.Int(1): u.String("A"),
			})),
		),
		dst: Obj(Save("v2", "x")),
		exp: arrObjStr("v2", "A"),
	},
	{
		name: "no var",
		inp:  arrObjInt("v", 1),
		src:  Obj(Has("v", u.Int(1))),
		dst:  Obj(Save("v2", "x")),
		err:  ErrVariableNotDefined,
	},
	{
		name: "var redeclared",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(2)),
		src: Obj(
			Save("v1", "x"),
			Save("v2", "x"),
		),
		dst: Obj(
			Save("v3", "x"),
			Save("v4", "x"),
		),
		err: ErrVariableRedeclared,
	},
	{
		name: "var val twice",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(1)),
		src: Obj(
			Save("v1", "x"),
			Save("v2", "x"),
		),
		dst: Obj(
			Save("v3", "x"),
		),
		exp: arrObjVal("v3", u.Int(1)),
	},
}

func TestOps(t *testing.T) {
	for _, c := range opCases {
		if c.exp == nil {
			c.exp = c.inp
		}
		t.Run(c.name, func(t *testing.T) {
			m := Map(c.src, c.dst)
			inp := c.inp()
			out, err := m.Do(inp)
			if c.err != nil {
				require.True(t, c.err.Is(err), "expected %v, got %v", c.err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.exp(), out)
			require.Equal(t, c.inp(), inp, "operation should clone the value")
			if c.noRev {
				return
			}
			m = m.Reverse()

			inp = c.exp()
			out, err = m.Do(inp)
			require.NoError(t, err)
			require.Equal(t, c.inp(), out)
			require.Equal(t, c.exp(), inp, "operation should clone the value")
		})
	}
}
