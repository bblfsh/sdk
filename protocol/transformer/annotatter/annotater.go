package annotatter

import (
	"gopkg.in/bblfsh/sdk.v0/protocol"
	"gopkg.in/bblfsh/sdk.v0/uast"
	"gopkg.in/bblfsh/sdk.v0/uast/ann"
)

type Annotatter struct {
	r *ann.Rule
}

func (t *Annotatter) Do(code string, e protocol.Encoding, n *uast.Node) error {
	return t.r.Apply(n)
}
