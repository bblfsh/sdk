package build

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/internal/docker"
	"github.com/bblfsh/sdk/v3/protocol"
	"google.golang.org/grpc"
	protocol1 "gopkg.in/bblfsh/sdk.v1/protocol"
)

const (
	cliPort      = "9432"
	dockerSchema = "docker-daemon:"
)

type ServerInstance struct {
	cli     *docker.Client
	user    *grpc.ClientConn
	bblfshd *docker.Container
}

func (d *ServerInstance) installFromDocker(ctx context.Context, lang, id string) error {
	if !strings.Contains(id, ":") {
		id += ":latest"
	}
	cmd := []string{"bblfshctl", "driver", "install", lang, dockerSchema + id}
	printCommand("docker", append([]string{"exec", id}, cmd...)...)
	e, err := d.cli.CreateExec(docker.CreateExecOptions{
		Context:      ctx,
		Container:    d.bblfshd.ID,
		AttachStdout: true, AttachStderr: true,
		Cmd: cmd,
	})
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	err = d.cli.StartExec(e.ID, docker.StartExecOptions{
		Context:      ctx,
		OutputStream: buf, ErrorStream: buf,
	})
	if err != nil {
		return err
	} else if str := buf.String(); strings.Contains(strings.ToLower(str), "error") {
		return errors.New(strings.TrimSpace(str))
	}
	return nil
}
func (d *ServerInstance) ClientV1(ctx context.Context) (protocol1.ProtocolServiceClient, error) {
	if d.user == nil {
		addr := d.bblfshd.NetworkSettings.IPAddress
		conn, err := grpc.DialContext(ctx, addr+":"+cliPort, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		d.user = conn
	}
	return protocol1.NewProtocolServiceClient(d.user), nil
}
func (d *ServerInstance) ClientV2(ctx context.Context) (driver.Driver, error) {
	if d.user == nil {
		addr := d.bblfshd.NetworkSettings.IPAddress
		conn, err := grpc.DialContext(ctx, addr+":"+cliPort, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		d.user = conn
	}
	return protocol.AsDriver(d.user), nil
}
func (s *ServerInstance) DumpLogs(w io.Writer) error {
	return getLogs(s.cli, s.bblfshd.ID, w)
}
func (d *ServerInstance) Close() error {
	if d.user != nil {
		_ = d.user.Close()
	}
	return d.cli.RemoveContainer(docker.RemoveContainerOptions{
		ID: d.bblfshd.ID, Force: true,
	})
}

// RunWithDriver starts a bblfshd server and installs a specified driver to it.
func RunWithDriver(bblfshdVers, lang, id string) (*ServerInstance, error) {
	cli, err := docker.Dial()
	if err != nil {
		return nil, err
	}
	const (
		bblfshd = "bblfsh/bblfshd"
		// needed to install driver from Docker instance
		sock = docker.Socket + ":" + docker.Socket
	)
	image := bblfshd
	if bblfshdVers != "" {
		image += ":" + bblfshdVers
	}

	printCommand("docker", "run", "--rm", "--privileged", "-v", sock, image)
	c, err := docker.Run(cli, docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: image,
		},
		HostConfig: &docker.HostConfig{
			AutoRemove: true,
			Privileged: true,
			Binds:      []string{sock},
		},
	})
	if err != nil {
		return nil, err
	}
	s := &ServerInstance{cli: cli, bblfshd: c}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	if err := s.installFromDocker(ctx, lang, id); err != nil {
		s.Close()
		return nil, err
	}
	return s, nil
}

func getLogs(cli *docker.Client, id string, w io.Writer) error {
	return cli.AttachToContainer(docker.AttachToContainerOptions{
		Container: id, OutputStream: w, ErrorStream: w,
		Logs: true, Stdout: true, Stderr: true,
	})
}
