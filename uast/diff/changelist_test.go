package diff

import (
    "fmt"
    "os"
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
    fd, err := os.Open(fmt.Sprintf("%v/config.txt", dataDir))
    require.NoError(t, err)
    var n int
    _, err = fmt.Fscanf(fd, "%d\n", &n)
    require.NoError(t, err)

    for i := 0; i < n; i++ {
        name := fmt.Sprintf("%v/testcase_%v", dataDir, i)
        t.Run(name, func(t *testing.T) {
            src_name := fmt.Sprintf("%v_src.uast", name)
            dst_name := fmt.Sprintf("%v_dst.uast", name)
            src := readUAST(t, src_name)
            dst := readUAST(t, dst_name)

            changes := Changes(src, dst)
            newsrc := Apply(src, changes)
            require.True(t, nodes.Equal(newsrc, dst))
        })
    }
}
