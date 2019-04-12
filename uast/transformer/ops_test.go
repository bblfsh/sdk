package transformer

import (
	"testing"

	u "github.com/bblfsh/sdk/v3/uast"
	un "github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/stretchr/testify/require"
	"gopkg.in/src-d/go-errors.v1"
)

func arrObjInt(key string, v int) func() un.Node {
	return arrObjVal(key, un.Int(v))
}

func arrObjStr(key string, v string) func() un.Node {
	return arrObjVal(key, un.String(v))
}

func arrObjVal(key string, v un.Value) func() un.Node {
	return func() un.Node {
		return un.Array{
			un.Object{key: v},
		}
	}
}

func arrObjVal2(key1, key2 string, v1, v2 un.Value) func() un.Node {
	return func() un.Node {
		return un.Array{
			un.Object{key1: v1, key2: v2},
		}
	}
}

var opsCases = []struct {
	name     string
	inp, exp func() un.Node
	src, dst Op
	err      *errors.Kind
	skip     bool
	noRev    bool // should only be set in exceptional cases
	expRev   func() un.Node
}{
	{
		name: "is",
		inp:  arrObjInt("v", 1),
		src:  Is(un.Int(1)),
		dst:  Is(un.Int(2)),
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
			"v": LookupVar("x", map[un.Value]un.Value{
				un.Int(1): un.String("A"),
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
		inp:  arrObjVal2("v1", "v2", un.Int(1), un.Int(2)),
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
		inp:  arrObjVal2("v1", "v2", un.Int(1), un.Int(1)),
		src: Obj{
			"v1": Var("x"),
			"v2": Var("x"),
		},
		dst: Obj{
			"v3": Var("x"),
		},
		exp: arrObjVal("v3", un.Int(1)),
	},
	{
		name: "partial transform",
		inp:  arrObjVal2("v1", "v2", un.Int(1), un.Int(2)),
		src: Part("other", Obj{
			"v1": Var("x"),
		}),
		dst: Part("other", Obj{
			"v3": Var("x"),
		}),
		exp: arrObjVal2("v3", "v2", un.Int(1), un.Int(2)),
	},
	{
		name: "unused field",
		inp:  arrObjVal2("v1", "v2", un.Int(1), un.Int(2)),
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
		inp:  arrObjVal("v1", un.Int(1)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[un.Value]Op{
				un.Int(1): String("a"),
				un.Int(2): String("b"),
			}),
		}),
		exp: arrObjVal2("v1", "v2", un.Int(1), un.String("a")),
	},
	{
		name: "op lookup 2",
		inp:  arrObjVal("v1", un.Int(2)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[un.Value]Op{
				un.Int(1): String("a"),
				un.Int(2): String("b"),
			}),
		}),
		exp: arrObjVal2("v1", "v2", un.Int(2), un.String("b")),
	},
	{
		name: "op lookup unhandled",
		inp:  arrObjVal("v1", un.Int(3)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[un.Value]Op{
				un.Int(1): String("a"),
				un.Int(2): String("b"),
			}),
		}),
		err: ErrUnhandledValueIn,
	},
	{
		name: "op lookup default",
		inp:  arrObjVal("v1", un.Int(3)),
		src: One(Obj{
			"v1": Var("x"),
		}),
		dst: One(Obj{
			"v1": Var("x"),
			"v2": LookupOpVar("x", map[un.Value]Op{
				un.Int(1): String("a"),
				un.Int(2): String("b"),
				nil:       String("c"),
			}),
		}),
		exp: arrObjVal2("v1", "v2", un.Int(3), un.String("c")),
	},
	{
		name: "op lookup order",
		inp:  arrObjVal2("v1", "v2", un.String("b"), un.Int(2)),
		src: One(Fields{
			{Name: "v2", Op: Var("x")},
			{Name: "v1", Op: LookupOpVar("x", map[un.Value]Op{
				un.Int(1): String("a"),
				un.Int(2): String("b"),
			})},
		}),
		dst: One(Obj{
			"v1": Var("x"),
		}),
		exp: arrObjVal("v1", un.Int(2)),
	},
	{
		name: "append",
		inp: func() un.Node {
			return un.Object{
				"t": un.Int(1),
			}
		},
		src: Obj{
			"t": Var("typ"),
		},
		dst: JoinObj(Fields{
			{Name: "t", Op: Var("typ")},
		}, Obj{
			"v2": Append(LookupOpVar("typ", map[un.Value]Op{
				un.Int(1): Arr(String("a")),
				un.Int(2): Arr(String("b")),
			}), Arr(String("c"), String("d"))),
		}),
		exp: func() un.Node {
			return un.Object{
				"t": un.Int(1),
				"v2": un.Array{
					un.String("a"),
					un.String("c"), un.String("d"),
				},
			}
		},
	},
	{
		name: "each",
		inp: func() un.Node {
			return un.Array{
				un.Object{"t": un.String("a"), "v": un.Int(1)},
				un.Object{"t": un.String("a"), "v": un.Int(2)},
				un.Object{"t": un.String("a"), "v": un.Int(3)},
			}
		},
		src: Each("objs", Part("part", Obj{
			"v": Var("val"),
		})),
		dst: Each("objs", Part("part", Obj{
			"v2": Var("val"),
		})),
		exp: func() un.Node {
			return un.Array{
				un.Object{"t": un.String("a"), "v2": un.Int(1)},
				un.Object{"t": un.String("a"), "v2": un.Int(2)},
				un.Object{"t": un.String("a"), "v2": un.Int(3)},
			}
		},
	},
	{
		name: "each (nil)",
		inp: func() un.Node {
			return nil
		},
		src: Each("objs", Var("part")),
	},
	{
		name: "field missing (fields)",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Op: String("a")},
		},
		err: ErrUnusedField,
	},
	{
		name: "field missing (obj)",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Obj{
			"t": String("a"),
		},
		err: ErrUnusedField,
	},
	{
		name: "join obj",
		inp: func() un.Node {
			return un.Object{
				"t":  un.String("a"),
				"v1": un.Int(1),
				"v2": un.Int(2),
			}
		},
		src: JoinObj(
			Obj{
				"t": String("a"),
			},
			Obj{
				"v1": Int(1),
				"v2": Int(2),
			},
		),
		dst: JoinObj(
			Obj{
				"t":  String("a"),
				"k1": Int(3),
			},
			Obj{
				"k2": Int(4),
			},
		),
		exp: func() un.Node {
			return un.Object{
				"t":  un.String("a"),
				"k1": un.Int(3),
				"k2": un.Int(4),
			}
		},
	},
	{
		name: "join obj unused",
		inp: func() un.Node {
			return un.Object{
				"t":  un.String("a"),
				"v1": un.Int(1),
				"v2": un.Int(2),
			}
		},
		src: JoinObj(
			Obj{
				"t": String("a"),
			},
			Obj{
				"v1": Int(1),
			},
		),
		err: ErrUnusedField,
	},
	{
		name: "join obj unused part",
		inp: func() un.Node {
			return un.Object{
				"t":  un.String("a"),
				"v1": un.Int(1),
				"v2": un.Int(2),
			}
		},
		src: JoinObj(
			Obj{
				"t": String("a"),
			},
			Part("part", Obj{
				"v1": Int(1),
			}),
		),
		dst: JoinObj(
			Obj{
				"t": String("a"),
				"m": Var("part"),
			},
			Obj{
				"k2": Int(4),
			},
		),
		exp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"m": un.Object{
					"v2": un.Int(2),
				},
				"k2": un.Int(4),
			}
		},
	},
	{
		name: "optional field",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
			}
		},
		src: Fields{
			{Name: "t", Op: String("a")},
			{Name: "v", Op: Var("val"), Optional: "exists"},
		},
	},
	{
		name: "optional nil field",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Op: String("a")},
			{Name: "v", Op: Opt("exists", Var("val"))},
		},
	},
	{
		name: "drop field any",
		inp: func() un.Node {
			return un.Object{
				"t": nil,
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Drop: true, Op: Any()},
			{Name: "v", Op: Is(nil)},
		},
		dst: Obj{"v": Is(nil)},
		exp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
	},
	{
		name: "drop field op",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Drop: true, Op: String("a")},
			{Name: "v", Op: Is(nil)},
		},
		dst: Obj{"v": Is(nil)},
		exp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
	},
	{
		name: "drop field op fail",
		inp: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Drop: true, Op: String("b")},
			{Name: "v", Op: Is(nil)},
		},
	},
	{
		name: "drop field missing any",
		inp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
		expRev: func() un.Node {
			return un.Object{
				"t": nil,
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Drop: true, Op: Any()},
			{Name: "v", Op: Is(nil)},
		},
		dst: Obj{"v": Is(nil)},
		exp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
	},
	{
		name: "drop field missing op",
		inp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
		expRev: func() un.Node {
			return un.Object{
				"t": un.String("a"),
				"v": nil,
			}
		},
		src: Fields{
			{Name: "t", Drop: true, Op: String("a")},
			{Name: "v", Op: Is(nil)},
		},
		dst: Obj{"v": Is(nil)},
		exp: func() un.Node {
			return un.Object{
				"v": nil,
			}
		},
	},
	{
		name: "roles field",
		inp: func() un.Node {
			return un.Object{
				u.KeyType: un.String("node"),
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
		exp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(1),
			}
		},
	},
	{
		name: "obj roles",
		inp: func() un.Node {
			return un.Object{
				u.KeyType: un.String("node"),
			}
		},
		src: ObjectRolesCustom("o", Obj{u.KeyType: String("node")}),
		dst: ObjectRolesCustom("o", Obj{u.KeyType: String("node")}, 1),
		exp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(1),
			}
		},
	},
	{
		name: "roles field exists",
		inp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
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
		exp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(2, 1),
			}
		},
	},
	{
		name: "roles field empty", skip: true, // TODO: track empty vs nil
		inp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(),
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
		exp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(1),
			}
		},
	},
	{
		name: "roles field nil", skip: true, // TODO: track empty vs nil
		inp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: nil,
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
		exp: func() un.Node {
			return un.Object{
				u.KeyType:  un.String("node"),
				u.KeyRoles: u.RoleList(1),
			}
		},
	},
	{
		name: "change type",
		inp: func() un.Node {
			return un.Object{
				u.KeyType: un.String("node"),
				"val":     un.String("a"),
			}
		},
		src: Check(
			Not(Has{"b": Any()}),
			Obj{
				u.KeyType: String("node"),
				"val":     Var("x"),
			},
		),
		dst: Arr(Var("x")),
		exp: func() un.Node {
			return un.Array{un.String("a")}
		},
	},
	{
		name: "quote",
		inp:  arrObjVal("v", un.String(`"a\"b"`)),
		src: Obj{
			"v": Quote(Var("x")),
		},
		dst: Obj{
			"v2": Var("x"),
		},
		exp: arrObjVal("v2", un.String(`a"b`)),
	},
	{
		name: "semantic",
		inp: func() un.Node {
			return un.Object{
				"type": un.String("ident"),
				"name": un.String("A"),
			}
		},
		src: Obj{
			"type": String("ident"),
			"name": Var("name"),
		},
		dst: UASTType(u.Identifier{}, Obj{
			"Name": Var("name"),
		}),
		exp: func() un.Node {
			return un.Object{
				u.KeyType: un.String("uast:Identifier"),
				"Name":    un.String("A"),
			}
		},
	},
	{
		name: "semantic (part)",
		inp: func() un.Node {
			return un.Object{
				u.KeyPos: un.Object{u.KeyStart: un.Int(1)},
				"type":   un.String("ident"),
				"name":   un.String("A"),
			}
		},
		src: Part("p", Obj{
			"type": String("ident"),
			"name": Var("name"),
		}),
		dst: UASTTypePart("p", u.Identifier{}, Obj{
			"Name": Var("name"),
		}),
		exp: func() un.Node {
			return un.Object{
				u.KeyPos:  un.Object{u.KeyStart: un.Int(1)},
				u.KeyType: un.String("uast:Identifier"),
				"Name":    un.String("A"),
			}
		},
	},
	{
		// TODO(dennwc): this will only work when FieldDesc will support negative checks
		skip: true,
		name: "has fields",
		inp: func() un.Node {
			return un.Array{
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(1)},
					"type":   un.String("ident"),
					"name":   un.String("A"),
				},
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(2)},
					"type":   un.String("ident"),
				},
			}
		},
		src: Part("p", CheckObj(
			HasFields{"name": false},
			Obj{
				"type": String("ident"),
			},
		)),
		dst: Part("p", Obj{
			"type": String("ident"),
			"name": String("<empty>"),
		}),
		exp: func() un.Node {
			return un.Array{
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(1)},
					"type":   un.String("ident"),
					"name":   un.String("A"),
				},
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(2)},
					"type":   un.String("ident"),
					"name":   un.String("<empty>"),
				},
			}
		},
	},
	{
		name: "has fields root",
		inp: func() un.Node {
			return un.Array{
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(1)},
					"type":   un.String("ident"),
					"name":   un.String("A"),
				},
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(2)},
					"type":   un.String("ident"),
				},
			}
		},
		src: CheckObj(
			HasFields{"name": false},
			Part("p", Obj{
				"type": String("ident"),
			}),
		),
		dst: Part("p", Obj{
			"type": String("ident"),
			"name": String("<empty>"),
		}),
		exp: func() un.Node {
			return un.Array{
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(1)},
					"type":   un.String("ident"),
					"name":   un.String("A"),
				},
				un.Object{
					u.KeyPos: un.Object{u.KeyStart: un.Int(2)},
					"type":   un.String("ident"),
					"name":   un.String("<empty>"),
				},
			}
		},
	},
	{
		name: "append complex",
		inp: func() un.Node {
			return un.Object{
				"type": un.String("A"),
				"arr": un.Array{
					un.Object{
						"type": un.String("B"),
						"name": un.String("B1"),
					},
				},
			}
		},
		src: Fields{
			{Name: "type", Op: String("A")},
			{
				Name: "arr", Optional: "hasArr",
				Op: Append(
					Each("elems",
						Obj{
							"type": String("B"),
							"name": Var("name"),
						},
					),
					Arr(
						Obj{
							"type": String("B"),
							"name": Var("a1_name"),
						},
					),
				),
			},
		},
		dst: Fields{
			{Name: "type", Op: String("A")},
			{
				Name: "arr", Optional: "hasArr",
				Op: Append(
					Each("elems",
						Obj{
							"type":  String("B"),
							"name2": Var("name"),
						},
					),
					Arr(
						Obj{
							"type":  String("B"),
							"name2": Var("a1_name"),
						},
					),
				),
			},
		},
		exp: func() un.Node {
			return un.Object{
				"type": un.String("A"),
				"arr": un.Array{
					un.Object{
						"type":  un.String("B"),
						"name2": un.String("B1"),
					},
				},
			}
		},
	},
	{
		name: "cases",
		inp: func() un.Node {
			return un.Array{
				un.Object{
					"type": un.String("ident"),
					"name": un.Array{un.String("A")},
				},
				un.Object{
					"type": un.String("ident"),
					"name": un.String("B"),
				},
			}
		},
		src: CasesObj("case",
			// common
			Obj{"type": String("ident")},
			Objects{
				// case 1: array
				Obj{"name": Arr(Var("name"))},
				// case 2: string
				Obj{"name": Check(OfKind(un.KindString), Var("name"))},
			},
		),
		dst: CasesObj("case",
			// common
			Obj{"type": String("ident")},
			Objects{
				// case 1: array
				Obj{
					"name2": Var("name"),
					"arr":   Bool(true),
				},
				// case 2: string
				Obj{
					"name2": Var("name"),
					"arr":   Bool(false),
				},
			},
		),
		exp: func() un.Node {
			return un.Array{
				un.Object{
					"type":  un.String("ident"),
					"name2": un.String("A"),
					"arr":   un.Bool(true),
				},
				un.Object{
					"type":  un.String("ident"),
					"name2": un.String("B"),
					"arr":   un.Bool(false),
				},
			}
		},
	},
}

func TestOps(t *testing.T) {
	for _, c := range opsCases {
		if c.exp == nil {
			c.exp = c.inp
		}
		if c.dst == nil {
			c.dst = c.src
		}
		t.Run(c.name, func(t *testing.T) {
			if c.skip {
				t.SkipNow()
			}
			m := Map(c.src, c.dst)

			do := func(m Mapping, er *errors.Kind, inpf, expf func() un.Node) bool {
				inp := inpf()
				out, err := Mappings(m).Do(inp)
				if er != nil {
					require.True(t, er.Is(err), "expected %v, got %v", er, err)
					return false
				}
				require.NoError(t, err)
				require.Equal(t, expf(), out, "transformation failed")
				require.Equal(t, inpf(), inp, "transformation should clone the value")
				return true
			}
			// test full transformation first
			if !do(m, c.err, c.inp, c.exp) {
				return // expected error case
			}
			if c.noRev {
				return
			}
			// test reverse transformation
			expRev := c.inp
			if c.expRev != nil {
				expRev = c.expRev
			}
			do(Reverse(m), nil, c.exp, expRev)

			// test identity transform (forward)
			m = Identity(c.src)
			do(m, nil, c.inp, expRev)

			// test identity transform (reverse)
			m = Identity(c.dst)
			do(m, nil, c.exp, c.exp)
		})
	}
}
