package build

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDriverInit(t *testing.T) {
	dir, err := ioutil.TempDir("", "driver_")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	const lang = "test"

	err = InitDriver(dir, lang, &InitOptions{
		Warning: func(format string, args ...interface{}) (int, error) {
			t.Logf(format, args...)
			return 0, nil
		},
		Notice: func(format string, args ...interface{}) (int, error) {
			t.Logf(format, args...)
			return 0, nil
		},
	})
	require.NoError(t, err)

	dir = filepath.Join(dir, lang+"-driver")

	err = UpdateSDK(dir, &UpdateOptions{
		DryRun: true,
		Warning: func(format string, args ...interface{}) (int, error) {
			t.Logf(format, args...)
			return 0, nil
		},
		Notice: func(format string, args ...interface{}) (int, error) {
			t.Logf(format, args...)
			return 0, nil
		},
	})
	require.NoError(t, err)
}
