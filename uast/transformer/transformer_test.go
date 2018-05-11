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
			}),
		),
		exp: u.Array{
			u.Object{
				u.KeyType: u.String("typed"),
				"p":       u.String("val1"),
				"key":     u.String("v"),
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
