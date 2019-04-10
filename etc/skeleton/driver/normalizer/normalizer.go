package normalizer

import (
	. "github.com/bblfsh/sdk/v3/uast/transformer"
	"github.com/bblfsh/sdk/v3/uast/transformer/positioner"
)

var Preprocess = Transformers([][]Transformer{
	{
		// ResponseMetadata is a transform that trims response metadata from AST.
		//
		// https://godoc.org/github.com/bblfsh/sdk/v3/uast#ResponseMetadata
		ResponseMetadata{
			TopLevelIsRootNode: false,
		},
	},
	{Mappings(Preprocessors...)},
}...)

var Normalize = Transformers([][]Transformer{

	{Mappings(Normalizers...)},
}...)

// Preprocessors is a block of AST preprocessing rules rules.
var Preprocessors = []Mapping{
	// ObjectToNode defines how to normalize common fields of native AST
	// (like node type, token, positional information).
	//
	// https://godoc.org/github.com/bblfsh/sdk/v3/uast#ObjectToNode
	ObjectToNode{
		InternalTypeKey: "...", // native AST type key name
	}.Mapping(),
}

// PreprocessCode is a preprocessor stage that can use the source code to
// fix tokens and positional information.
var PreprocessCode = []CodeTransformer{
	positioner.FromOffset(),
}

// Normalizers is the main block of normalization rules to convert native AST to semantic UAST.
var Normalizers = []Mapping{}
