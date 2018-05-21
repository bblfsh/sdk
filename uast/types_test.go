package uast

import (
	"testing"

	"reflect"

	"github.com/stretchr/testify/require"
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
	exp  Node
}{
	{
		name: "Position",
		obj:  Position{Offset: 5, Line: 2, Col: 3},
		exp: Object{
			KeyType:  String("uast:Position"),
			"offset": Int(5),
			"line":   Int(2),
			"col":    Int(3),
		},
	},
	{
		name: "Position (consts)",
		obj:  Position{Offset: 0, Line: 0, Col: 0},
		exp: Object{
			KeyType:    String(TypePosition),
			KeyPosOff:  Int(0),
			KeyPosLine: Int(0),
			KeyPosCol:  Int(0),
		},
	},
	{
		name: "Positions",
		obj: Positions{
			KeyStart: {Offset: 3, Line: 2, Col: 1},
			KeyEnd:   {Offset: 4, Line: 2, Col: 2},
		},
		exp: Object{
			KeyType: String(TypePositions),
			KeyStart: Object{
				KeyType:    String(TypePosition),
				KeyPosOff:  Int(3),
				KeyPosLine: Int(2),
				KeyPosCol:  Int(1),
			},
			KeyEnd: Object{
				KeyType:    String(TypePosition),
				KeyPosOff:  Int(4),
				KeyPosLine: Int(2),
				KeyPosCol:  Int(2),
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
