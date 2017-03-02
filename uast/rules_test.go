package uast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOnInternalType(t *testing.T) {
	require := require.New(t)

	require.True(OnInternalType("foo")(
		&Node{InternalType: "foo"},
	))
	require.False(OnInternalType("foo")(
		&Node{InternalType: "other"},
	))

	require.True(OnInternalType("foo")(
		&Node{InternalType: "other"},
		&Node{InternalType: "foo"},
	))

	require.True(OnInternalType("foo", "bar")(
		&Node{InternalType: "foo"},
		&Node{InternalType: "bar"},
	))
	require.True(OnInternalType("foo", "bar")(
		&Node{InternalType: "other"},
		&Node{InternalType: "foo"},
		&Node{InternalType: "bar"},
	))
	require.False(OnInternalType("foo", "bar")(
		&Node{InternalType: "foo"},
		&Node{InternalType: "other"},
		&Node{InternalType: "bar"},
	))
}

func TestOnPath(t *testing.T) {
	require := require.New(t)

	require.True(OnPath(OnInternalType("foo"))(
		&Node{InternalType: "foo"},
	))
	require.False(OnPath(OnInternalType("foo"))(
		&Node{InternalType: "other"},
	))

	require.True(OnPath(OnInternalType("foo"), OnInternalType("bar"))(
		&Node{InternalType: "foo"},
		&Node{InternalType: "bar"},
	))
	require.False(OnPath(OnInternalType("foo"), OnInternalType("bar"))(
		&Node{InternalType: "foo"},
		&Node{InternalType: "other"},
		&Node{InternalType: "bar"},
	))
}

func ExampleOnPath() {
	s := OnPath(OnInternalType("A"))
	r := s(&Node{InternalType: "A"})
	fmt.Println(`OnPath(OnInternalType("A")) matches A =`, r)

	r = s(&Node{InternalType: "B"}, &Node{InternalType: "A"})
	fmt.Println(`OnPath(OnInternalType("A")) matches B A =`, r)

	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnPath(OnInternalType("A")) matches A B =`, r)

	s = OnPath(OnInternalType("A"), OnInternalType("B"))
	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnPath(OnInternalType("A"), OnInternalType("B")) matches A B =`, r)

	r = s(&Node{InternalType: "X"}, &Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnPath(OnInternalType("A"), OnInternalType("B")) matches X A B =`, r)

	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"}, &Node{InternalType: "X"})
	fmt.Println(`OnPath(OnInternalType("A"), OnInternalType("B")) matches A B X =`, r)

	//Output:
	// OnPath(OnInternalType("A")) matches A = true
	// OnPath(OnInternalType("A")) matches B A = true
	// OnPath(OnInternalType("A")) matches A B = false
	// OnPath(OnInternalType("A"), OnInternalType("B")) matches A B = true
	// OnPath(OnInternalType("A"), OnInternalType("B")) matches X A B = true
	// OnPath(OnInternalType("A"), OnInternalType("B")) matches A B X = false
}

func ExampleOnInternalType() {
	s := OnInternalType("A")
	r := s(&Node{InternalType: "A"})
	fmt.Println(`OnInternalType("A") matches A =`, r)

	r = s(&Node{InternalType: "B"}, &Node{InternalType: "A"})
	fmt.Println(`OnInternalType("A") matches B A =`, r)

	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnInternalType("A") matches A B =`, r)

	s = OnInternalType("A", "B")
	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnInternalType("A", "B") matches A B =`, r)

	r = s(&Node{InternalType: "X"}, &Node{InternalType: "A"}, &Node{InternalType: "B"})
	fmt.Println(`OnInternalType("A", "B") matches X A B =`, r)

	r = s(&Node{InternalType: "A"}, &Node{InternalType: "B"}, &Node{InternalType: "X"})
	fmt.Println(`OnInternalType("A", "B") matches A B X =`, r)

	//Output:
	// OnInternalType("A") matches A = true
	// OnInternalType("A") matches B A = true
	// OnInternalType("A") matches A B = false
	// OnInternalType("A", "B") matches A B = true
	// OnInternalType("A", "B") matches X A B = true
	// OnInternalType("A", "B") matches A B X = false
}
