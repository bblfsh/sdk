package uast

import (
	"testing"

	"reflect"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

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

var casesToNode = []struct {
	name string
	obj  interface{}
	exp  nodes.Node
}{
	{
		name: "Position",
		obj:  Position{Offset: 5, Line: 2, Col: 3},
		exp: nodes.Object{
			KeyType:  nodes.String("uast:Position"),
			"offset": nodes.Int(5),
			"line":   nodes.Int(2),
			"col":    nodes.Int(3),
		},
	},
	{
		name: "Position (consts)",
		obj:  Position{Offset: 0, Line: 0, Col: 0},
		exp: nodes.Object{
			KeyType:    nodes.String(TypePosition),
			KeyPosOff:  nodes.Int(0),
			KeyPosLine: nodes.Int(0),
			KeyPosCol:  nodes.Int(0),
		},
	},
	{
		name: "Positions",
		obj: Positions{
			KeyStart: {Offset: 3, Line: 2, Col: 1},
			KeyEnd:   {Offset: 4, Line: 2, Col: 2},
		},
		exp: nodes.Object{
			KeyType: nodes.String(TypePositions),
			KeyStart: nodes.Object{
				KeyType:    nodes.String(TypePosition),
				KeyPosOff:  nodes.Int(3),
				KeyPosLine: nodes.Int(2),
				KeyPosCol:  nodes.Int(1),
			},
			KeyEnd: nodes.Object{
				KeyType:    nodes.String(TypePosition),
				KeyPosOff:  nodes.Int(4),
				KeyPosLine: nodes.Int(2),
				KeyPosCol:  nodes.Int(2),
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
			err = nodeAs(got, nv)
			require.NoError(t, err)
			require.Equal(t, c.obj, nv.Interface())
		})
	}
}
