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
		src:  Obj{"v": Int(1)},
		dst:  Obj{"v2": Int(2)},
		exp:  arrObjInt("v2", 2),
	},
	{
		name: "has nil",
		inp:  arrObjVal("v", nil),
		src:  Obj{"v": Is(nil)},
		dst:  Obj{"v2": Int(2)},
		exp:  arrObjInt("v2", 2),
	},
	{
		name: "obj save",
		inp:  arrObjInt("v", 1),
		src:  Obj{"v": Var("x")},
		dst:  Obj{"v2": Var("x")},
		exp:  arrObjInt("v2", 1),
	},
	{
		name: "save nil",
		inp:  arrObjVal("v", nil),
		src:  Obj{"v": Var("x")},
		dst:  Obj{"v2": Var("x")},
		exp:  arrObjVal("v2", nil),
	},
	{
		name: "arr save",
		inp:  arrObjInt("v", 1),
		src:  One(Obj{"v": Var("x")}),
		dst:  One(Obj{"v2": Var("x")}),
		exp:  arrObjInt("v2", 1),
	},
	{
		name: "lookup save",
		inp:  arrObjInt("v", 1),
		src: Obj{
			"v": LookupVar("x", map[u.Value]u.Value{
				u.Int(1): u.String("A"),
			}),
		},
		dst: Obj{"v2": Var("x")},
		exp: arrObjStr("v2", "A"),
	},
	{
		name: "no var",
		inp:  arrObjInt("v", 1),
		src:  Obj{"v": Int(1)},
		dst:  Obj{"v2": Var("x")},
		err:  ErrVariableNotDefined,
	},
	{
		name: "var redeclared",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(2)),
		src: Obj{
			"v1": Var("x"),
			"v2": Var("x"),
		},
		dst: Obj{
			"v3": Var("x"),
			"v4": Var("x"),
		},
		err: ErrVariableRedeclared,
	},
	{
		name: "var val twice",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(1)),
		src: Obj{
			"v1": Var("x"),
			"v2": Var("x"),
		},
		dst: Obj{
			"v3": Var("x"),
		},
		exp: arrObjVal("v3", u.Int(1)),
	},
	{
		name: "partial transform",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(2)),
		src: Part("other", Obj{
			"v1": Var("x"),
		}),
		dst: Part("other", Obj{
			"v3": Var("x"),
		}),
		exp: arrObjVal2("v3", "v2", u.Int(1), u.Int(2)),
	},
	{
		name: "unused field",
		inp:  arrObjVal2("v1", "v2", u.Int(1), u.Int(2)),
		src: Obj{
			"v1": Var("x"),
		},
		dst: Obj{
			"v3": Var("x"),
		},
		err: ErrUnusedField,
	},
	{
		name: "op lookup 1",
		inp:  arrObjVal("v1", u.Int(1)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[u.Value]Op{
				u.Int(1): String("a"),
				u.Int(2): String("b"),
			}),
		}),
		exp: arrObjVal2("v1", "v2", u.Int(1), u.String("a")),
	},
	{
		name: "op lookup 2",
		inp:  arrObjVal("v1", u.Int(2)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[u.Value]Op{
				u.Int(1): String("a"),
				u.Int(2): String("b"),
			}),
		}),
		exp: arrObjVal2("v1", "v2", u.Int(2), u.String("b")),
	},
	{
		name: "op lookup unhandled",
		inp:  arrObjVal("v1", u.Int(3)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[u.Value]Op{
				u.Int(1): String("a"),
				u.Int(2): String("b"),
			}),
		}),
		err: ErrUnhandledValueIn,
	},
	{
		name: "op lookup order",
		inp:  arrObjVal2("v1", "v2", u.String("b"), u.Int(2)),
		src: One(Fields{
			{Name: "v2", Op: Var("x")},
			{Name: "v1", Op: LookupOpVar("x", map[u.Value]Op{
				u.Int(1): String("a"),
				u.Int(2): String("b"),
			})},
		}),
		dst: One(Obj{
			"v1": Var("x"),
		}),
		exp: arrObjVal("v1", u.Int(2)),
	},
	{
		name: "append",
		inp: func() u.Node {
			return u.Object{
				"t": u.Int(1),
			}
		},
		src: Obj{
			"t": Var("typ"),
		},
		dst: Pre(Fields{
			{Name: "t", Op: Var("typ")},
		}, Obj{
			"v2": Append(LookupOpVar("typ", map[u.Value]Op{
				u.Int(1): Arr(String("a")),
				u.Int(2): Arr(String("b")),
			}), Arr(String("c"), String("d"))),
		}),
		exp: func() u.Node {
			return u.Object{
				"t": u.Int(1),
				"v2": u.List{
					u.String("a"),
					u.String("c"), u.String("d"),
				},
			}
		},
	},
	{
		name: "each",
		inp: func() u.Node {
			return u.List{
				u.Object{"t": u.String("a"), "v": u.Int(1)},
				u.Object{"t": u.String("a"), "v": u.Int(2)},
				u.Object{"t": u.String("a"), "v": u.Int(3)},
			}
		},
		src: Each("objs", Part("part", Obj{
			"v": Var("val"),
		})),
		dst: Each("objs", Part("part", Obj{
			"v2": Var("val"),
		})),
		exp: func() u.Node {
			return u.List{
				u.Object{"t": u.String("a"), "v2": u.Int(1)},
				u.Object{"t": u.String("a"), "v2": u.Int(2)},
				u.Object{"t": u.String("a"), "v2": u.Int(3)},
			}
		},
	},
	{
		name: "optional field",
		inp: func() u.Node {
			return u.Object{
				"t": u.String("a"),
			}
		},
		src: Fields{
			{Name: "t", Op: String("a")},
			{Name: "v", Op: Var("val"), Optional: "exists"},
		},
	},
	{
		name: "roles field",
		inp: func() u.Node {
			return u.Object{
				u.KeyType: u.String("node"),
			}
		},
		src: Fields{
			{Name: u.KeyType, Op: String("node")},
			RolesField("roles"),
		},
		dst: Fields{
			{Name: u.KeyType, Op: String("node")},
			RolesField("roles", 1),
		},
		exp: func() u.Node {
			return u.Object{
				u.KeyType:  u.String("node"),
				u.KeyRoles: u.RoleList(1),
			}
		},
	},
	{
		name: "roles field exists",
		inp: func() u.Node {
			return u.Object{
				u.KeyType:  u.String("node"),
				u.KeyRoles: u.RoleList(2),
			}
		},
		src: Fields{
			{Name: u.KeyType, Op: String("node")},
			RolesField("roles"),
		},
		dst: Fields{
			{Name: u.KeyType, Op: String("node")},
			RolesField("roles", 1),
		},
		exp: func() u.Node {
			return u.Object{
				u.KeyType:  u.String("node"),
				u.KeyRoles: u.RoleList(2, 1),
			}
		},
	},
}

func TestOps(t *testing.T) {
	for _, c := range opCases {
		if c.exp == nil {
			c.exp = c.inp
		}
		t.Run(c.name, func(t *testing.T) {
			if c.dst == nil {
				c.dst = c.src
			}
			m := Map("test", c.src, c.dst)
			inp := c.inp()
			out, err := m.Do(inp)
			if c.err != nil {
				require.True(t, c.err.Is(err), "expected %v, got %v", c.err, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, c.exp(), out, "forward transformation failed")
			require.Equal(t, c.inp(), inp, "forward transformation should clone the value")
			if c.noRev {
				return
			}
			m = m.Reverse()

			inp = c.exp()
			out, err = m.Do(inp)
			require.NoError(t, err)
			require.Equal(t, c.inp(), out, "reverse transformation failed")
			require.Equal(t, c.exp(), inp, "reverse transformation should clone the value")
		})
	}
}
