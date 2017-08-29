package normalizer

import (
	"github.com/bblfsh/sdk/protocol/driver"
	"github.com/bblfsh/sdk/protocol/native"
)

// ParserBuilder creates a parser that transform source code files into *uast.Node.
func ParserBuilder(opts driver.ParserOptions) (parser driver.Parser, err error) {
	psr, err := native.ExecParser(ToNoder, opts.NativeBin)
	if err != nil {
		return psr, err
	}

	switch ToNoder.PositionFill {
	case native.None:
		parser = psr
	case native.OffsetFromLineCol:
		parser = &driver.TransformationParser{
			Parser:         psr,
			Transformation: driver.FillOffsetFromLineCol,
		}
	case native.LineColFromOffset:
		parser = &driver.TransformationParser{
			Parser:         psr,
			Transformation: driver.FillLineColFromOffset,
		}
	}

	return parser, nil
}
