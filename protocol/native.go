package protocol

import (
	"io"
	"os/exec"

	"github.com/bblfsh/sdk/protocol/jsonlines"
)

// ParseNativeASTRequest to use with the native AST parser. This is for internal
// use.
type ParseNativeASTRequest struct {
	Content string
}

// ParseNativeASTResponse is the reply to ParseASTRequest by the native AST
// parser.
type ParseNativeASTResponse struct {
	Status Status
	Errors []string
	AST    interface{}
}

// NativeClient is a wrapper of the native command.
type NativeClient struct {
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	closer io.Closer
	cmd    *exec.Cmd
}

// ExecNative executes the given command and returns a *NativeClient for it.
func ExecNative(path string, args ...string) (*NativeClient, error) {
	cmd := exec.Command(path, args...)
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

	return &NativeClient{
		enc:    jsonlines.NewEncoder(in),
		dec:    jsonlines.NewDecoder(out),
		closer: in,
		cmd:    cmd,
	}, nil
}

// ParseNativeAST sends a request to the native client and returns its response.
func (c *NativeClient) ParseNativeAST(req *ParseNativeASTRequest) (*ParseNativeASTResponse, error) {
	resp := &ParseNativeASTResponse{}
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
