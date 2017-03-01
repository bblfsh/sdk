package protocol

import (
	"encoding/json"
	"io"
	"os/exec"
)

const (
	// Ok status code.
	Ok = "ok"
	// Error status code. It is replied when the driver has got the AST with errors.
	Error = "error"
	// Fatal status code. It is replied when the driver hasn't could get the AST.
	Fatal = "fatal"
	// ParseAST is the Action identifier to parse an AST.
	ParseAST = "ParseAST"
)

// Request is the message the driver receives.
//proteus:generate
type Request struct {
	Action          string `codec:"action"`
	Language        string `codec:"language,omitempty"`
	LanguageVersion string `codec:"language_version,omitempty"`
	Content         string `codec:"content"`
}

// Response is the replied message.
//proteus:generate
type Response struct {
	Status          string      `codec:"status"`
	Errors          []string    `codec:"errors,omitempty"`
	Driver          string      `codec:"driver"`
	Language        string      `codec:"language"`
	LanguageVersion string      `codec:"language_version"`
	AST             interface{} `codec:"ast"`
}

// NativeClient is a wrapper of the native command.
type NativeClient struct {
	enc    *json.Encoder
	dec    *json.Decoder
	closer io.Closer
	cmd    *exec.Cmd
}

// ExecNative executes the given command and returns a *NativeClient for it.
func ExecNative(path string) (*NativeClient, error) {
	cmd := exec.Command(path)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &NativeClient{json.NewEncoder(in), json.NewDecoder(out), in, cmd}, nil
}

// Request sends a request to the native client and returns its response.
func (c *NativeClient) Request(req *Request) (*Response, error) {
	resp := &Response{}
	_ = c.enc.Encode(req)
	return resp, c.dec.Decode(resp)
}

// Close closes the native client.
func (c *NativeClient) Close() error {
	if err := c.closer.Close(); err != nil {
		return err
	}

	return c.cmd.Wait()
}
