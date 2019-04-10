package normalizer

import (
	"github.com/bblfsh/sdk/v3/uast/role"
	. "github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/bblfsh/sdk/v3/uast/transformer/positioner"
)

// Native is the of list `transformer.Transformer` to apply to a native AST.
// To learn more about the Transformers and the available ones take a look to:
// https://godoc.org/github.com/bblfsh/sdk/v3/uast/transformer
var Native = Transformers([][]Transformer{
	// The main block of transformation rules.
	{Mappings(Annotations...)},
	{
		// RolesDedup is used to remove duplicate roles assigned by multiple
		// transformation rules.
		RolesDedup(),
	},
}...)

// Code is a special block of transformations that are applied at the end
// and can access original source code file. It can be used to improve or
// fix positional information.
//
// https://godoc.org/github.com/bblfsh/sdk/v3/uast/transformer/positioner
var Code = []CodeTransformer{
	positioner.FromOffset(),
}

// Annotations is a list of individual transformations to annotate a native AST with roles.
var Annotations = []Mapping{
	AnnotateType("internal-type", nil, role.Incomplete),
}
