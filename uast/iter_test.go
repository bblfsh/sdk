package uast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOrderPathIter(t *testing.T) {
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

	iter := NewOrderPathIter(NewPath(n))
	var result []string
	for {
		p := iter.Next()
		if p.IsEmpty() {
			break
		}

		result = append(result, p.Node().(Object).Type())
	}

	require.Equal([]string{"a", "aa", "ab", "aba", "ac"}, result)
}
