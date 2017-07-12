package driver

import (
	"io"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/jsonlines"
)

// Client is a client to communicate with a driver.
type Client struct {
	closer io.Closer
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
}

func NewClient(in io.WriteCloser, out io.Reader) *Client {
	return &Client{
		closer: in,
		enc:    jsonlines.NewEncoder(in),
		dec:    jsonlines.NewDecoder(out),
	}
}

func (c *Client) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	if err := c.enc.Encode(req); err != nil {
		return newFatalResponse(err)
	}

	resp := &protocol.ParseResponse{}
	if err := c.dec.Decode(resp); err != nil {
		return newFatalResponse(err)
	}

	return resp
}

func (c *Client) Close() error {
	return c.closer.Close()
}
