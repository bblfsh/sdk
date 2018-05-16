package uast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	fixtureDir = "fixtures"
)

func TestClone(t *testing.T) {
	o1 := Object{"v": Int(1)}
	o2 := Object{"k": o1, "v2": Int(2)}
	arr := make(Array, 0, 3)
	arr = append(arr, o1, o2)

	arr2 := arr.Clone()

	o1["new"] = Int(0)
	o2["new"] = Int(0)
	arr[0] = Int(3)

	require.Equal(t, Array{
		Object{"v": Int(1)},
		Object{"k": Object{"v": Int(1)}, "v2": Int(2)},
	}, arr2)

	require.Equal(t, Array{
		Int(3),
		Object{
			"k": Object{
				"v":   Int(1),
				"new": Int(0),
			},
			"v2":  Int(2),
			"new": Int(0),
		},
	}, arr)
}

func TestWalkPreOrder(t *testing.T) {
	require := require.New(t)

	n := Object{
		KeyType: String("a"),
		"a":     Object{KeyType: String("aa")},
		"b": Object{
			KeyType: String("ab"),
			"a":     Object{KeyType: String("aba")},
		},
		"c": Object{KeyType: String("ac")},
	}

	var result []string
	WalkPreOrder(n, func(n Node) bool {
		if obj, ok := n.(Object); ok {
			result = append(result, obj.Type())
		}
		return true
	})

	require.Equal([]string{"a", "aa", "ab", "aba", "ac"}, result)
}

func TestApply(t *testing.T) {
	o1 := Object{"v": Int(1)}
	o2 := Object{"k": o1, "v": Int(2)}
	arr := Array{o1, o2}

	out, ok := Apply(arr, func(n Node) (Node, bool) {
		switch n := n.(type) {
		case Object:
			n = n.CloneObject()
			v, _ := n["v"].(Int)
			n["v2"] = v
			return n, true
		case Array:
			n[0] = Int(3)
			return n, true
		case Int:
			n++
			return n, true
		}
		return n, false
	})
	require.True(t, ok)
	require.Equal(t, Array{
		Int(3),
		Object{
			"k": Object{
				"v":  Int(2),
				"v2": Int(2),
			},
			"v":  Int(3),
			"v2": Int(3),
		},
	}, out)
}
