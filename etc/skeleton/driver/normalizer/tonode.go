package normalizer

import (
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
)

// ToNode is an instance of `uast.ObjectToNode`, defining how to transform an
// into a UAST (`uast.Node`).
//
// https://godoc.org/gopkg.in/bblfsh/sdk.v1/uast#ObjectToNode
var ToNode = &transformer.ObjectToNode{}
