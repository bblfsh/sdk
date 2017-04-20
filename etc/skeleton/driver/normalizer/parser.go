package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

// ASTParserBuilder creates a parser that transform source code files into *uast.Node.
func ASTParserBuilder(opts driver.ASTParserOptions) (driver.ASTParser, error) {
	toNoder := &native.ObjectToNoder{}
	parser, err := native.ExecParser(toNoder, opts.NativeBin)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
