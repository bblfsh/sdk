package protocol

import (
	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"
)

// UASTClient encapsulates a *NativeClient to provide ParseUAST.
type UASTClient struct {
	*NativeClient
	ToNoder  uast.ToNoder
	Annotate *ann.Rule
}

// ParseUAST processes a *ParseUASTRequest and returns a *ParseUASTResponse.
func (h *UASTClient) ParseUAST(req *ParseUASTRequest) (*ParseUASTResponse, error) {
	resp, err := h.ParseNativeAST(&ParseNativeASTRequest{
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}

	var node *uast.Node
	if resp.Status != Fatal {
		node, err = h.ToNoder.ToNode(resp.AST)
		if err != nil {
			return nil, err
		}

		if err := h.Annotate.Apply(node); err != nil {
			return nil, err
		}
	}

	uastResp := &ParseUASTResponse{
		Status: resp.Status,
		Errors: resp.Errors,
		UAST:   node,
	}

	return uastResp, nil
}
