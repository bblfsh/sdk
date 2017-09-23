package driver

import (
	"time"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
)

// Driver implements a driver.
type Driver struct {
	NativeDriver

	// Version of the driver.
	Version string
	// Build identifier.
	Build string
	// Commit hash.
	Commit string
	// ObjectToNode
	ObjectToNode uast.ObjectToNode
	// Transformers list of transformer to apply to the the UAST after parse it.
	Transformers []transformer.Tranformer
}

// Parse process a protocol.ParseRequest, calling to the native driver. At this
// point the language from the request is ignored.
func (d *Driver) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	start := time.Now()
	nr := d.ParseNative(&protocol.ParseNativeRequest{
		Content:  req.Content,
		Encoding: req.Encoding,
	})

	r := &protocol.ParseResponse{
		Status: nr.Status,
		Errors: nr.Errors,
	}

	defer func() {
		r.Elapsed = time.Since(start)
	}()

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
