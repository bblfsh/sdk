package native

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testExecClient() (*Client, error) {
	return ExecClient("go", "run", "../internal/testnative/main.go")
}

func TestExecClient(t *testing.T) {
	require := require.New(t)

	n, err := testExecClient()
	require.NoError(err)
	require.NotNil(n)

	resp, err := n.ParseNativeAST(&ParseASTRequest{
		Content: "foo",
	})
	require.NoError(err)
	require.NotNil(resp)

	err = n.Close()
	require.NoError(err)
}

func TestExecClientBadPath(t *testing.T) {
	require := require.New(t)

	n, err := ExecClient("non-existent")
	require.Error(err)
	require.Nil(n)
}

func TestExecClientMalfunctioning(t *testing.T) {
	require := require.New(t)

	n, err := ExecClient("echo")
	require.NoError(err)
	require.NotNil(n)

	_, err = n.ParseNativeAST(&ParseASTRequest{
		Content: "foo",
	})

	require.Error(err)
}

func TestExecClientMalformed(t *testing.T) {
	require := require.New(t)

	n, err := ExecClient("yes")
	require.NoError(err)
	require.NotNil(n)

	_, err = n.ParseNativeAST(&ParseASTRequest{
		Content: "foo",
	})

	require.Error(err)
}
