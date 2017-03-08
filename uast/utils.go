package uast

// Tokens returns a slice of tokens contained in the node.
func Tokens(n *Node) []string {
	var tokens []string
	iter := NewPreOrderPathIter(NewPath(n))
	for {
		p := iter.Next()
		if p.IsEmpty() {
			break
		}

		n := p.Node()
		if n.Token != "" {
			tokens = append(tokens, n.Token)
		}
	}

	return tokens
}
