package transformer

import (
	"testing"

	"github.com/stretchr/testify/require"
	u "gopkg.in/bblfsh/sdk.v2/uast"
)

var mappingCases = []struct {
	name     string
	skip     bool
	inp, exp u.Node
	m        Transformer
}{
	{
		name: "trim meta",
		inp: u.Object{
			"the_root": u.Object{
				"k": u.String("v"),
			},
		},
		m: ResponseMetadata{
			TopLevelIsRootNode: false,
		},
		exp: u.Object{
			"k": u.String("v"),
		},
	},
	{
		name: "leave meta",
		inp: u.Object{
			"the_root": u.Object{
				"k": u.String("v"),
			},
		},
		m: ResponseMetadata{
			TopLevelIsRootNode: true,
		},
	},
	{
		name: "roles dedup",
		inp: u.Array{
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(1, 2, 1),
			},
		},
		m: RolesDedup(),
		exp: u.Array{
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(1, 2),
			},
		},
	},
	{
		name: "typed and generic",
		inp: u.Array{
			u.Object{
				u.KeyType: u.String("typed"),
				"pred":    u.String("val1"),
				"k":       u.String("v"),
			},
			u.Object{
				"pred": u.String("val2"),
				"k":    u.String("v"),
			},
			u.Object{
				"pred2": u.String("val3"),
			},
		},
		m: Mappings(
			Map("test",
				Part("_", Obj{
					"pred": Var("x"),
				}),
				Part("_", Obj{
					"p": Var("x"),
				}),
			),
			MapAST("typed", Obj{
				"k": Var("x"),
			}, Obj{
				"key": Var("x"),
			}, 10),
		),
		exp: u.Array{
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(10),
				"p":        u.String("val1"),
				"key":      u.String("v"),
			},
			u.Object{
				"p": u.String("val2"),
				"k": u.String("v"),
			},
			u.Object{
				"pred2": u.String("val3"),
			},
		},
	},
	{
		name: "annotate no roles",
		inp: u.Array{
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(1),
				"pred":     u.String("val1"),
			},
			u.Object{
				u.KeyType: u.String("typed"),
				"pred":    u.String("val2"),
			},
		},
		m: AnnotateIfNoRoles("typed", 10),
		exp: u.Array{
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(1),
				"pred":     u.String("val1"),
			},
			u.Object{
				u.KeyType:  u.String("typed"),
				u.KeyRoles: u.RoleList(10),
				"pred":     u.String("val2"),
			},
		},
	},
}

func TestMappings(t *testing.T) {
	for _, c := range mappingCases {
		if c.exp == nil {
			c.exp = c.inp
		}
		t.Run(c.name, func(t *testing.T) {
			out, err := c.m.Do(c.inp)
			require.NoError(t, err)
			require.Equal(t, c.exp, out, "transformation failed")
		})
	}
}
