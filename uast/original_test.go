package uast

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	fixtureDir = "fixtures"
)

func TestOriginalToNode(t *testing.T) {
	require := require.New(t)

	f, err := getFixture("java_example_1.json")
	require.NoError(err)

	c := &BaseOriginalToNoder{}
	n, err := c.OriginalToNode(f)
	require.NoError(err)
	require.NotNil(n)
	fmt.Println("NODE", n)
}

func getFixture(name string) (map[string]interface{}, error) {
	path := filepath.Join(fixtureDir, name)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(f)
	data := map[string]interface{}{}
	if err := d.Decode(&data); err != nil {
		_ = f.Close()
		return nil, err
	}

	if err := f.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
