package diff

import (
    "strings"
    "os"
    "path/filepath"
    "io/ioutil"
    "testing"
    "gopkg.in/bblfsh/sdk.v2/uast/yaml"
    "gopkg.in/bblfsh/sdk.v2/uast/nodes"
    "github.com/stretchr/testify/require"
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
                src_name := filepath.Join(dataDir, name + "_src.uast")
                dst_name := filepath.Join(dataDir, name + "_dst.uast")
                src := readUAST(t, src_name)
                dst := readUAST(t, dst_name)

                changes := Changes(src, dst)
                newsrc := changes.Apply(src)
                require.True(t, nodes.Equal(newsrc, dst))
            })
        }
    }
}
