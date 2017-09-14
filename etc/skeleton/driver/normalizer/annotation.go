package normalizer

import (
	"gopkg.in/bblfsh/sdk.v1/uast"
	. "gopkg.in/bblfsh/sdk.v1/uast/ann"
)

// AnnotationRules annotate a UAST with roles.
var AnnotationRules = On(Any).Roles(uast.File)
