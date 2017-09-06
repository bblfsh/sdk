package normalizer

import (
	. "gopkg.in/bblfsh/sdk.v0/uast"
	. "gopkg.in/bblfsh/sdk.v0/uast/ann"
)

// AnnotationRules annotate a UAST with roles.
var AnnotationRules = On(Any).Roles(File)
