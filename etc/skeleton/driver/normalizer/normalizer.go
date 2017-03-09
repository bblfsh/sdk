package normalizer

import (
	. "github.com/bblfsh/sdk/uast"
	. "github.com/bblfsh/sdk/uast/ann"
)

// NativeToNoder implement a ToNoder to convert from the native AST to a *Node.
// BaseToNoder can be used (with parameters) for most cases.
var NativeToNoder = &BaseToNoder{}

// AnnotationRules annotate a UAST with roles.
var AnnotationRules = On(Any).Roles(File)
