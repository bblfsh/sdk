package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

// ParserBuilder creates a parser that transform source code files into *uast.Node.
func ParserBuilder(opts driver.ParserOptions) (driver.Parser, error) {
	toNoder := &native.ObjectToNoder{}
	parser, err := native.ExecParser(toNoder, opts.NativeBin)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
