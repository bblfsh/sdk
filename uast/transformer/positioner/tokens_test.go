package positioner

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

func newPos(start, end int) nodes.Node {
	return uast.Positions{
		uast.KeyStart: {
			Offset: uint32(start),
			Line:   1,
			Col:    uint32(1 + start),
		},
		uast.KeyEnd: {
			Offset: uint32(end),
			Line:   1,
			Col:    uint32(1 + end),
		},
	}.ToObject()
}

func TestTokenFromSource(t *testing.T) {
	testAst := nodes.Array{
		nodes.Object{
			uast.KeyType: nodes.String("test:Var"),
			uast.KeyPos:  newPos(1, 13),
			"Names": nodes.Array{
				nodes.Object{
					uast.KeyType:  nodes.String("test:Ident"),
					uast.KeyToken: nodes.String(""),
					uast.KeyPos:   newPos(5, 6),
				},
				nodes.Object{
					uast.KeyType:  nodes.String("test:Ident"),
					uast.KeyToken: nodes.String(""),
					uast.KeyPos:   newPos(8, 9),
				},
			},
			"Type": nodes.Object{
				uast.KeyType:  nodes.String("test:Type"),
				uast.KeyToken: nodes.String(""),
				uast.KeyPos:   newPos(10, 13),
				"Notes":       nodes.String(""),
			},
		},
	}

	var cases = []struct {
		name     string
		source   string
		ast, exp nodes.Node
		conf     TokenFromSource
	}{
		{
			name:   "fix all",
			source: " var a, b int",
			conf:   TokenFromSource{}, // defaults
			ast:    testAst,
			exp: nodes.Array{
				nodes.Object{
					uast.KeyType: nodes.String("test:Var"),
					uast.KeyPos:  newPos(1, 13),
					"Names": nodes.Array{
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String("a"),
							uast.KeyPos:   newPos(5, 6),
						},
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String("b"),
							uast.KeyPos:   newPos(8, 9),
						},
					},
					"Type": nodes.Object{
						uast.KeyType:  nodes.String("test:Type"),
						uast.KeyToken: nodes.String("int"),
						uast.KeyPos:   newPos(10, 13),
						"Notes":       nodes.String(""),
					},
				},
			},
		},
		{
			name:   "only type",
			source: " var a, b int",
			conf: TokenFromSource{
				Types: []string{"test:Type"},
			},
			ast: testAst,
			exp: nodes.Array{
				nodes.Object{
					uast.KeyType: nodes.String("test:Var"),
					uast.KeyPos:  newPos(1, 13),
					"Names": nodes.Array{
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String(""),
							uast.KeyPos:   newPos(5, 6),
						},
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String(""),
							uast.KeyPos:   newPos(8, 9),
						},
					},
					"Type": nodes.Object{
						uast.KeyType:  nodes.String("test:Type"),
						uast.KeyToken: nodes.String("int"),
						uast.KeyPos:   newPos(10, 13),
						"Notes":       nodes.String(""),
					},
				},
			},
		},
		{
			name:   "specific field",
			source: " var a, b int",
			conf: TokenFromSource{
				Key: "Notes",
			},
			ast: testAst,
			exp: nodes.Array{
				nodes.Object{
					uast.KeyType: nodes.String("test:Var"),
					uast.KeyPos:  newPos(1, 13),
					"Names": nodes.Array{
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String(""),
							uast.KeyPos:   newPos(5, 6),
						},
						nodes.Object{
							uast.KeyType:  nodes.String("test:Ident"),
							uast.KeyToken: nodes.String(""),
							uast.KeyPos:   newPos(8, 9),
						},
					},
					"Type": nodes.Object{
						uast.KeyType:  nodes.String("test:Type"),
						uast.KeyToken: nodes.String(""),
						uast.KeyPos:   newPos(10, 13),
						"Notes":       nodes.String("int"),
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root := c.ast
			tr := c.conf.OnCode(c.source)
			got, err := tr.Do(root.Clone())
			require.NoError(t, err)
			require.Equal(t, c.exp, got)
		})
	}
}
