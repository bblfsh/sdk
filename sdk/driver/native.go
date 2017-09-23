package driver

import (
	"io"
	"os"
	"os/exec"
	"time"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/jsonlines"
)

// NativeBin default location of the native driver binary.
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

// Start executes the given native driver and prepares it to parse code.
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

// NativeParse sends a request to the native driver and returns its response.
func (d *NativeDriver) NativeParse(req *protocol.NativeParseRequest) *protocol.NativeParseResponse {
	start := time.Now()
	resp := &protocol.ParseNativeResponse{}

	_ = d.enc.Encode(req)
	if err := d.dec.Decode(resp); err != nil {
		resp = &protocol.ParseNativeResponse{
			Status: protocol.Fatal,
			Errors: []string{err.Error()},
		}
	}

	resp.Elapsed = time.Since(start)
	return resp
}

// Stop stops the execution of the native driver.
func (d *NativeDriver) Stop() error {
	if err := d.closer.Close(); err != nil {
		return err
	}

	return d.cmd.Wait()
}
