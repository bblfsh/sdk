package main

import (
	"io"
	"os"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/jsonlines"
)

type ParseASTResponse struct {
	Status protocol.Status
	Errors []string
	AST    interface{}
}

func main() {
	dec := jsonlines.NewDecoder(os.Stdin)
	enc := jsonlines.NewEncoder(os.Stdout)
	for {
		req := &protocol.ParseASTRequest{}
		if err := dec.Decode(req); err != nil {
			if err == io.EOF {
				os.Exit(0)
			}

			os.Exit(-1)
		}

		resp := &ParseASTResponse{
			Status: protocol.Ok,
			AST: map[string]interface{}{
				"root": map[string]interface{}{
					"key": "val",
				},
			},
		}
		if err := enc.Encode(resp); err != nil {
			os.Exit(-1)
		}
	}
}
