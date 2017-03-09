package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testExecNative() (*NativeClient, error) {
	return ExecNative("go", "run", "internal/testnative/main.go")
}

func TestExecNative(t *testing.T) {
	require := require.New(t)

	n, err := testExecNative()
	require.NoError(err)
	require.NotNil(n)

	resp, err := n.ParseNativeAST(&ParseNativeASTRequest{
		Content: "foo",
	})
	require.NoError(err)
	require.NotNil(resp)

	err = n.Close()
	require.NoError(err)
}

func TestExecNativeBadPath(t *testing.T) {
	require := require.New(t)

	n, err := ExecNative("non-existent")
	require.Error(err)
	require.Nil(n)
}

func TestExecNativeMalfunctioning(t *testing.T) {
	require := require.New(t)

	n, err := ExecNative("echo")
	require.NoError(err)
	require.NotNil(n)

	_, err = n.ParseNativeAST(&ParseNativeASTRequest{
		Content: "foo",
	})

	require.Error(err)
}

func TestExecNativeMalformed(t *testing.T) {
	require := require.New(t)

	n, err := ExecNative("yes")
	require.NoError(err)
	require.NotNil(n)

	_, err = n.ParseNativeAST(&ParseNativeASTRequest{
		Content: "foo",
	})

	require.Error(err)
}
