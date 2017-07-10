// native package is used to parse AST using an external binary.
package native

import (
	"io"
	"os"
	"os/exec"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/jsonlines"
	"github.com/bblfsh/sdk/uast"
)

// ToNoder transforms a decoded JSON into a *uast.Node. This decoded JSON can be
// any type based on maps, slices, int, float64 and strings.
type ToNoder interface {
	// ToNode transforms an arbitrary value into a *Node, or emits an error.
	ToNode(interface{}) (*uast.Node, error)
}

// Parser uses a *Client to parse source code and a ToNoder to convert it to a
// *uast.Node.
type Parser struct {
	Client  *Client
	ToNoder ToNoder
}

// ExecParser constructs a uast.Parser based on a native parser binary.
func ExecParser(toNoder ToNoder, path string, args ...string) (*Parser, error) {
	c, err := ExecClient(path, args...)
	if err != nil {
		return nil, err
	}

	return &Parser{Client: c, ToNoder: toNoder}, nil
}

func (p *Parser) ParseUAST(req *protocol.ParseUASTRequest) *protocol.ParseUASTResponse {
	nativeResp, err := p.Client.ParseNative(&ParseNativeRequest{
		Content:  req.Content,
		Encoding: req.Encoding,
	})

	if err != nil {
		return &protocol.ParseUASTResponse{
			Status: protocol.Fatal,
			Errors: []string{err.Error()},
		}
	}

	resp := &protocol.ParseUASTResponse{
		Status: nativeResp.Status,
		Errors: nativeResp.Errors,
	}

	if resp.Status == protocol.Fatal {
		return resp
	}

	uast, err := p.ToNoder.ToNode(nativeResp.AST)
	if err != nil {
		resp.Status = protocol.Fatal
		resp.Errors = append(resp.Errors, err.Error())
		return resp
	}

	resp.UAST = uast
	return resp
}

func (p *Parser) Close() error {
	return p.Client.Close()
}

// ParseNativeRequest to use with the native AST parser. This is for internal use.
type ParseNativeRequest struct {
	Content  string            `json:"content"`
	Encoding protocol.Encoding `json:"encoding"`
}

// ParseNativeResponse is the reply to ParseNativeRequest by the native AST parser.
type ParseNativeResponse struct {
	Status protocol.Status `json:"status"`
	Errors []string        `json:"errors"`
	AST    interface{}     `json:"ast"`
}

// Client is a wrapper of the native command.
type Client struct {
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	closer io.Closer
	cmd    *exec.Cmd
}

// ExecNative executes the given command and returns a *Client for it.
func ExecClient(path string, args ...string) (*Client, error) {
	cmd := exec.Command(path, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	errReader, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go io.Copy(os.Stderr, errReader)

	return &Client{
		enc:    jsonlines.NewEncoder(in),
		dec:    jsonlines.NewDecoder(out),
		closer: in,
		cmd:    cmd,
	}, nil
}

// ParseNative sends a request to the native client and returns its response.
func (c *Client) ParseNative(req *ParseNativeRequest) (*ParseNativeResponse, error) {
	resp := &ParseNativeResponse{}
	_ = c.enc.Encode(req)
	return resp, c.dec.Decode(resp)
}

// Close closes the native client.
func (c *Client) Close() error {
	if err := c.closer.Close(); err != nil {
		return err
	}

	return c.cmd.Wait()
}
