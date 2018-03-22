package transformer

import "strconv"

func Quote(op Op) Op {
	return StringConv(op, func(s string) (string, error) {
		return strconv.Quote(s), nil
	}, strconv.Unquote)
}
