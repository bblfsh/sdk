package uast

// Tokens returns a slice of tokens contained in the node.
func Tokens(n *Node) []string {
	var tokens []string
	iter := NewOrderPathIter(NewPath(n))
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

// CyclomaticComplexity returns the cyclomatic complexity for the node. This uses the method of
// counting one + one of the following UAST Roles if present on any children:
// If | SwitchCase |  SwitchDefault | For[Each] | [Do]While | TryCatch | Continue | OpBoolean* | Goto
// Important: since some languages allow for code defined
// outside function definitions, this won't check that the Node has the role FunctionDeclarationRole
// so the user should check that if the intended use is calculating the complexity of a function/method.
// If the children contain more than one function definitions, the value will not be averaged between
// the total number of function declarations but given as a total.
//
// Original paper: http://www.literateprogramming.com/mccabe.pdf
//
// Some practical implementations counting tokens in the code. They sometimes differ; for example
// some of them count the switch "default" as an incrementor, some consider all return values minus the
// last, some of them consider "else" (which is wrong IMHO, but not for elifs, remember than the IfElse
// token in the UAST is really an Else not an "else if", elseifs would have a children If token), some
// consider throw and finally while others only the catch, etc.
//
// GMetrics: http://gmetrics.sourceforge.net/gmetrics-CyclomaticComplexityMetric.html
// Go: https://github.com/fzipp/gocyclo/blob/master/gocyclo.go#L214
// SonarQube (include rules for many languages): https://docs.sonarqube.org/display/SONAR/Metrics+-+Complexity
func CyclomaticComplexity(n *Node)  int {
	complx := 1

	iter := NewOrderPathIter(NewPath(n))

	for {
		p := iter.Next()
		if p.IsEmpty() {
			break
		}
		n := p.Node()
		for _, r := range n.Roles {
			switch(r) {
			case If, SwitchCase, SwitchDefault, For, ForEach, While,
			     DoWhile, TryCatch, Continue, OpBooleanAnd, OpBooleanOr,
			     OpBooleanXor:
				complx++
			}
		}
	}
	return complx
}
