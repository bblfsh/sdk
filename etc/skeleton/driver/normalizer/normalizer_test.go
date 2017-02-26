package normalizer

import (
	"os/exec"
	"testing"
	"time"

	"github.com/bblfsh/sdk"

	"github.com/stretchr/testify/require"
)

func TestNativeBinary(t *testing.T) {
	r := require.New(t)

	cmd := exec.Command(sdk.NativeBinTest)
	err := cmd.Start()
	r.Nil(err)

	time.Sleep(time.Second)
	err = cmd.Process.Kill()
	r.Nil(err)
}
