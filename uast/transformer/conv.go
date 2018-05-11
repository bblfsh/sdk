package transformer

import "strconv"

// Quote uses strconv.Quote/Unquote to wrap provided string value.
func Quote(op Op) Op {
	return StringConv(op, strconv.Unquote, func(s string) (string, error) {
		return strconv.Quote(s), nil
	})
}
