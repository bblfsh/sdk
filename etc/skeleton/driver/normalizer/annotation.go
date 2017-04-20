package normalizer

import (
	. "github.com/bblfsh/sdk/uast"
	. "github.com/bblfsh/sdk/uast/ann"
)

// AnnotationRules annotate a UAST with roles.
var AnnotationRules = On(Any).Roles(File)
