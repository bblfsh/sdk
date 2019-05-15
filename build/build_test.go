package build

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/internal/docker"
	"github.com/bblfsh/sdk/v3/protocol"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

func TestIsApkOrApt(t *testing.T) {
	for _, c := range []struct {
		cmd string
		exp bool
	}{
		{`echo`, false},
		{`apk add xxx`, true},
		{`apt install xxx`, true},
		{`apt-get install xxx`, true},
		{`apt update && apt install xxx`, true},
		{`apt update && echo`, false},
		{`apt update || echo`, false},
		{`apt update ; echo`, false},
		{`apt install x || apt install y`, true},
		{`apt install x; apt install y`, true},
	} {
		t.Run(c.cmd, func(t *testing.T) {
			if c.exp != isApkOrApt(c.cmd) {
				t.Fatal("should be", c.exp)
			}
		})
	}
}

func TestDriverBuildAndRun(t *testing.T) {
	// If changing any skeleton files, make sure to 'make bindata' before running this.
	const lang = "test"

	dir, err := ioutil.TempDir("", "driver_")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	err = InitDriver(dir, "test", nil)
	require.NoError(t, err)

	dir = filepath.Join(dir, "test-driver")

	sdkRoot, err := filepath.Abs("../")
	require.NoError(t, err)

	// replace the SDK with the dev version, so we can actually test our changes
	out, err := exec.Command(
		"go", "mod", "edit",
		"-replace=github.com/bblfsh/sdk/v3="+sdkRoot,
		filepath.Join(dir, "go.mod"),
	).CombinedOutput()
	if err != nil {
		require.Fail(t, "mod edit failed", "%s", string(out))
	}

	b, err := NewDriver(dir)
	require.NoError(t, err)

	id, err := b.Build("")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	d := &dockerDriver{ctx: ctx, img: id}
	err = d.Start()
	require.NoError(t, err)
	defer d.Close()

	const src = "foo"
	ast, err := d.Parse(ctx, src, &driver.ParseOptions{
		Mode:     driver.ModeNative,
		Language: lang,
	})
	require.NoError(t, err)
	require.Equal(t, nodes.Object{
		"Encoding": nodes.String("utf8"),
		"content":  nodes.String(src),
	}, ast)
}

var (
	_ driver.Driver = (*dockerDriver)(nil)
	_ driver.Module = (*dockerDriver)(nil)
)

type dockerDriver struct {
	ctx context.Context
	img string

	cli  *docker.Client
	cont *docker.Container
	cc   *grpc.ClientConn
	d    driver.Driver
}

func (d *dockerDriver) Start() error {
	cli, err := docker.Dial()
	if err != nil {
		return err
	}
	cont, err := docker.Run(cli, docker.CreateContainerOptions{
		Context: d.ctx,
		Config: &docker.Config{
			Image: d.img,
		},
		HostConfig: &docker.HostConfig{
			AutoRemove: true,
		},
	})
	if err != nil {
		return err
	}
	d.cli = cli
	d.cont = cont

	cc, err := grpc.DialContext(d.ctx, cont.NetworkSettings.IPAddress+":9432",
		grpc.WithInsecure(), grpc.WithBlock(),
	)
	if err != nil {
		_ = d.Close()
		return err
	}
	d.cc = cc
	d.d = protocol.AsDriver(cc)
	return nil
}

func (d *dockerDriver) Close() error {
	if d.cc != nil {
		_ = d.cc.Close()
	}
	return d.cli.RemoveContainer(docker.RemoveContainerOptions{
		ID:            d.cont.ID,
		RemoveVolumes: true,
		Force:         true,
	})
}

func (d *dockerDriver) Parse(ctx context.Context, src string, opts *driver.ParseOptions) (nodes.Node, error) {
	return d.d.Parse(ctx, src, opts)
}

func (d *dockerDriver) Version(ctx context.Context) (driver.Version, error) {
	return d.d.Version(ctx)
}

func (d *dockerDriver) Languages(ctx context.Context) ([]manifest.Manifest, error) {
	return d.d.Languages(ctx)
}
