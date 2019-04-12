package uast_test

import (
	"reflect"
	"testing"

	. "github.com/bblfsh/sdk/v3/uast"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/stretchr/testify/require"
)

func init() {
	RegisterPackage("test", arrayNode{})
}

var casesTypeOf = []struct {
	name string
	typ  interface{}
	exp  string
}{
	{
		name: "Position",
		typ:  Position{},
		exp:  "uast:Position",
	},
}

func TestTypeOf(t *testing.T) {
	for _, c := range casesTypeOf {
		t.Run(c.name, func(t *testing.T) {
			require.Equal(t, c.exp, TypeOf(c.typ))
		})
	}
}

func expPos(off int, line int, col int) nodes.Object {
	return nodes.Object{
		KeyType:    nodes.String(TypePosition),
		KeyPosOff:  nodes.Uint(off),
		KeyPosLine: nodes.Uint(line),
		KeyPosCol:  nodes.Uint(col),
	}
}

type arrayNode struct {
	GenNode
	Array []Any `json:"Array"`
}

var casesToNode = []struct {
	name   string
	obj    interface{}
	exp    nodes.Node
	expObj interface{}
}{
	{
		name: "Position",
		obj:  Position{Offset: 5, Line: 2, Col: 3},
		exp: nodes.Object{
			KeyType:  nodes.String("uast:Position"),
			"offset": nodes.Uint(5),
			"line":   nodes.Uint(2),
			"col":    nodes.Uint(3),
		},
	},
	{
		name: "Position (consts)",
		obj:  Position{Offset: 0, Line: 0, Col: 0},
		exp:  expPos(0, 0, 0),
	},
	{
		name: "Positions",
		obj: Positions{
			KeyStart: {Offset: 3, Line: 2, Col: 1},
			KeyEnd:   {Offset: 4, Line: 2, Col: 2},
		},
		exp: nodes.Object{
			KeyType:  nodes.String(TypePositions),
			KeyStart: expPos(3, 2, 1),
			KeyEnd:   expPos(4, 2, 2),
		},
	},
	{
		name: "Bool",
		obj: Bool{
			GenNode: GenNode{
				Positions: Positions{
					KeyStart: {Offset: 3, Line: 2, Col: 1},
					KeyEnd:   {Offset: 8, Line: 2, Col: 6},
				},
			},
			Value: true,
		},
		exp: nodes.Object{
			KeyType: nodes.String("uast:Bool"),
			KeyPos: nodes.Object{
				KeyType:  nodes.String(TypePositions),
				KeyStart: expPos(3, 2, 1),
				KeyEnd:   expPos(8, 2, 6),
			},
			"Value": nodes.Bool(true),
		},
	},
	{
		name: "Alias",
		obj: Alias{
			GenNode: GenNode{
				Positions: Positions{
					KeyStart: {Offset: 3, Line: 2, Col: 1},
					KeyEnd:   {Offset: 8, Line: 2, Col: 6},
				},
			},
			Name: Identifier{
				GenNode: GenNode{
					Positions: Positions{
						KeyStart: {Offset: 3, Line: 2, Col: 1},
						KeyEnd:   {Offset: 4, Line: 2, Col: 2},
					},
				},
				Name: "ok",
			},
			Node: String{
				GenNode: GenNode{
					Positions: Positions{
						KeyStart: {Offset: 6, Line: 2, Col: 4},
						KeyEnd:   {Offset: 8, Line: 2, Col: 6},
					},
				},
				Value: "A",
			},
		},
		exp: nodes.Object{
			KeyType: nodes.String("uast:Alias"),
			KeyPos: nodes.Object{
				KeyType:  nodes.String(TypePositions),
				KeyStart: expPos(3, 2, 1),
				KeyEnd:   expPos(8, 2, 6),
			},
			"Name": nodes.Object{
				KeyType: nodes.String("uast:Identifier"),
				KeyPos: nodes.Object{
					KeyType:  nodes.String(TypePositions),
					KeyStart: expPos(3, 2, 1),
					KeyEnd:   expPos(4, 2, 2),
				},
				"Name": nodes.String("ok"),
			},
			"Node": nodes.Object{
				KeyType: nodes.String("uast:String"),
				KeyPos: nodes.Object{
					KeyType:  nodes.String(TypePositions),
					KeyStart: expPos(6, 2, 4),
					KeyEnd:   expPos(8, 2, 6),
				},
				"Value": nodes.String("A"), "Format": nodes.String(""),
			},
		},
		expObj: Alias{
			GenNode: GenNode{
				Positions: Positions{
					KeyStart: {Offset: 3, Line: 2, Col: 1},
					KeyEnd:   {Offset: 8, Line: 2, Col: 6},
				},
			},
			Name: Identifier{
				GenNode: GenNode{
					Positions: Positions{
						KeyStart: {Offset: 3, Line: 2, Col: 1},
						KeyEnd:   {Offset: 4, Line: 2, Col: 2},
					},
				},
				Name: "ok",
			},
			Node: nodes.Object{
				KeyType: nodes.String("uast:String"),
				KeyPos: nodes.Object{
					KeyType:  nodes.String(TypePositions),
					KeyStart: expPos(6, 2, 4),
					KeyEnd:   expPos(8, 2, 6),
				},
				"Value": nodes.String("A"), "Format": nodes.String(""),
			},
		},
	},
	{
		name: "arrayNode",
		obj: arrayNode{
			GenNode: GenNode{
				Positions: Positions{
					KeyStart: {Offset: 3, Line: 2, Col: 1},
					KeyEnd:   {Offset: 8, Line: 2, Col: 6},
				},
			},
			Array: []Any{
				Identifier{Name: "a", GenNode: GenNode{
					Positions: Positions{},
				}},
				String{Value: "a", GenNode: GenNode{
					Positions: Positions{},
				}},
			},
		},
		exp: nodes.Object{
			KeyType: nodes.String("test:arrayNode"),
			KeyPos: nodes.Object{
				KeyType:  nodes.String(TypePositions),
				KeyStart: expPos(3, 2, 1),
				KeyEnd:   expPos(8, 2, 6),
			},
			"Array": nodes.Array{
				nodes.Object{
					KeyType: nodes.String("uast:Identifier"),
					KeyPos:  nodes.Object{KeyType: nodes.String(TypePositions)},
					"Name":  nodes.String("a"),
				},
				nodes.Object{
					KeyType:  nodes.String("uast:String"),
					KeyPos:   nodes.Object{KeyType: nodes.String(TypePositions)},
					"Value":  nodes.String("a"),
					"Format": nodes.String(""),
				},
			},
		},
	},
}

func TestToNode(t *testing.T) {
	for _, c := range casesToNode {
		t.Run(c.name, func(t *testing.T) {
			got, err := ToNode(c.obj)
			require.NoError(t, err)
			require.Equal(t, c.exp, got)

			nv := reflect.New(reflect.TypeOf(c.obj)).Elem()
			err = NodeAs(got, nv)
			require.NoError(t, err)
			expObj := c.expObj
			if expObj == nil {
				expObj = c.obj
			}
			require.Equal(t, expObj, nv.Interface())
		})
	}
}
