package uast

import (
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
