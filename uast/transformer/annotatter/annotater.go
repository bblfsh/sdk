package annotatter

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/ann"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
)

var _ transformer.Transformer = Annotatter{}

type Annotatter struct {
	r *ann.Rule
}

func NewAnnotatter(r *ann.Rule) Annotatter {
	return Annotatter{r: r}
}

func (t Annotatter) Do(n uast.Node) (uast.Node, error) {
	err := t.r.Apply(n)
	return n, err
}
