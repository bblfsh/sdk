package driver

import (
	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/protocol/transformer"
	"gopkg.in/bblfsh/sdk.v1/uast"
)

// Driver implements a driver.
type Driver struct {
	NativeDriver

	// Version of the driver.
	Version string
	// Build identifier.
	Build string
	// ObjectToNode
	ObjectToNode uast.ObjectToNoder
	// Transformers list of transformer to apply to the the UAST after parse it.
	Transformers []transformer.Tranformer
}

func (d *Driver) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	nr := d.ParseNative(&protocol.ParseNativeRequest{
		Content:  req.Content,
		Encoding: req.Encoding,
	})

	r := &protocol.ParseResponse{
		Status: nr.Status,
		Errors: nr.Errors,
	}

	if r.Status == protocol.Fatal {
		return r
	}

	n, err := d.ObjectToNode.ToNode(nr.AST)
	if err != nil {
		r.Status = protocol.Fatal
		r.Errors = append(r.Errors, err.Error())
		return r
	}

	for _, t := range d.Transformers {
		if err := t.Do(req.Content, req.Encoding, n); err != nil {
			return nil
		}
	}

	r.UAST = n
	return r
}
