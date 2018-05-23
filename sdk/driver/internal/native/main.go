package main

import (
	"io"
	"os"

	"gopkg.in/bblfsh/sdk.v2/protocol"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
	"gopkg.in/bblfsh/sdk.v2/sdk/jsonlines"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

func main() {
	dec := jsonlines.NewDecoder(os.Stdin)
	enc := jsonlines.NewEncoder(os.Stdout)
	for {
		req := &driver.InternalParseRequest{}
		if err := dec.Decode(req); err != nil {
			if err == io.EOF {
				os.Exit(0)
			}

			if err := enc.Encode(&driver.InternalParseResponse{
				Status: driver.Status(protocol.Fatal),
				Errors: []string{err.Error()},
			}); err != nil {
				os.Exit(-1)
			}

			continue
		}

		resp := &driver.InternalParseResponse{
			Status: driver.Status(protocol.Ok),
			AST: nodes.Object{
				"root": nodes.Object{
					"key": nodes.String("val"),
				},
			},
		}

		if err := enc.Encode(resp); err != nil {
			os.Exit(-1)
		}
	}
}
