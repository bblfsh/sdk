package driver

import (
	"io"
	"os"
	"os/exec"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/protocol/jsonlines"
)

const NativeBin = "/opt/driver/bin/native"

// NativeDriver is a wrapper of the native command.
type NativeDriver struct {
	// Binary path to the location of the native driver binary.
	Binary string
	// Args argument to pass to the native bianary if needed.
	Args []string

	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	closer io.Closer
	cmd    *exec.Cmd
}

// NewNativeDriver executes the given command and returns a *NativeDriver for it.
func (d *NativeDriver) Start() error {
	if d.Binary == "" {
		d.Binary = NativeBin
	}

	d.cmd = exec.Command(d.Binary, d.Args...)
	stdin, err := d.cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	d.enc = jsonlines.NewEncoder(stdin)
	d.dec = jsonlines.NewDecoder(stdout)
	d.closer = stdin

	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)

	return d.cmd.Start()
}

// ParseNative sends a request to the native NativeDriver and returns its response.
func (d *NativeDriver) ParseNative(req *protocol.ParseNativeRequest) *protocol.ParseNativeResponse {
	resp := &protocol.ParseNativeResponse{}
	_ = d.enc.Encode(req)

	if err := d.dec.Decode(resp); err != nil {
		return &protocol.ParseNativeResponse{
			Status: protocol.Fatal,
			Errors: []string{err.Error()},
		}
	}

	return resp
}

// Close closes the native NativeDriver.
func (d *NativeDriver) Stop() error {
	if err := d.closer.Close(); err != nil {
		return err
	}

	return d.cmd.Wait()
}
