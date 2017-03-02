package uast

// Tokens returns a slice of tokens contained in the node.
func Tokens(n *Node) []string {
	var tokens []string
	err := PreOrderVisit(n, func(path ...*Node) error {
		n := path[len(path)-1]
		if n.Token != "" {
			tokens = append(tokens, n.Token)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return tokens
}
