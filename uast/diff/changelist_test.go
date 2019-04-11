package diff

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	uastyml "gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

const dataDir = "./testdata"

func readUAST(t testing.TB, path string) nodes.Node {
	data, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	nd, err := uastyml.Unmarshal(data)
	require.NoError(t, err)
	return nd
}

func TestChangelist(t *testing.T) {
	dir, err := os.Open(dataDir)
	require.NoError(t, err)
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	require.NoError(t, err)

	for _, fname := range names {
		if strings.HasSuffix(fname, "_src.uast") {
			name := fname[:len(fname)-len("_src.uast")]

			t.Run(name, func(t *testing.T) {
				srcName := filepath.Join(dataDir, name+"_src.uast")
				dstName := filepath.Join(dataDir, name+"_dst.uast")
				src := readUAST(t, srcName)
				dst := readUAST(t, dstName)

				changes := Changes(src, dst)
				newsrc, err := changes.Apply(src)
				require.NoError(t, err)
				require.True(t, nodes.Equal(newsrc, dst))
			})
		}
	}
}
