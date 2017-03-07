package normalizer

import (
	"github.com/bblfsh/sdk/uast"
)

// Implement a uast.ToNoder to convert from the native AST to a *uast.Node.
// uast.BaseToNoder can be used (with parameters) for most cases.
var ToNoder = &uast.BaseToNoder{}

// Annotate annotates a *uast.Node with roles.
func Annotate(n *uast.Node) error {
	return nil
}
