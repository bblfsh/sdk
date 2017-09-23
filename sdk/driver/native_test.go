package driver

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v1/protocol"

	"github.com/stretchr/testify/require"
)

func TestNativeDriverParser(t *testing.T) {
	require := require.New(t)

	d := &NativeDriver{
		Binary: "internal/native/mock",
	}
	err := d.Start()
	require.NoError(err)

	r := d.ParseNative(&protocol.ParseNativeRequest{
		Content: "foo",
	})

	require.NotNil(r)
	require.Equal(r.Status, protocol.Ok)

	err = d.Stop()
	require.NoError(err)
}

func TestNativeDriverStart_BadPath(t *testing.T) {
	require := require.New(t)

	d := &NativeDriver{Binary: "non-existent"}
	err := d.Start()
	require.Error(err)
}

func TestNativeDriverParser_Malfunctioning(t *testing.T) {
	require := require.New(t)

	d := &NativeDriver{
		Binary: "echo",
	}

	err := d.Start()
	require.Nil(err)

	r := d.ParseNative(&protocol.ParseNativeRequest{
		Content: "foo",
	})

	require.Equal(r.Status, protocol.Fatal)
	require.Equal(len(r.Errors), 1)
}

func TestNativeDriverParser_Malformed(t *testing.T) {
	require := require.New(t)

	d := &NativeDriver{
		Binary: "yes",
	}

	err := d.Start()
	require.NoError(err)

	r := d.ParseNative(&protocol.ParseNativeRequest{
		Content: "foo",
	})

	require.Equal(r.Status, protocol.Fatal)
	require.Equal(len(r.Errors), 1)
}
