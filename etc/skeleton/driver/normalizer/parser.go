package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

// UASTParserBuilder creates a parser that transform source code files into *uast.Node.
func UASTParserBuilder(opts driver.UASTParserOptions) (driver.UASTParser, error) {
	toNoder := &native.ObjectToNoder{}
	parser, err := native.ExecParser(toNoder, opts.NativeBin)
	if err != nil {
		return nil, err
	}

	return parser, nil
}
