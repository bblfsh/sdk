package uast

// Tokens returns a slice of tokens contained in the node.
func Tokens(n *Node) []string {
	var tokens []string
	iter := NewPreOrderIter(n)
	for {
		n := iter.Next()
		if n == nil {
			break
		}

		if n.Token != "" {
			tokens = append(tokens, n.Token)
		}
	}

	return tokens
}
