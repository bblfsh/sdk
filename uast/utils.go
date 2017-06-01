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

// CyclomaticComplexity returns the cyclomatic complexity for the node. The cyclomatic complexity
// is a quantitative measure of the number of linearly independent paths through a program's source code.
// It was developed by Thomas J. McCabe, Sr. in 1976. For a formal description see:
// https://en.wikipedia.org/wiki/Cyclomatic_complexity
// And the original paper: http://www.literateprogramming.com/mccabe.pdf

// This implementation uses PMD implementation as reference and uses the method of
// counting one + one of the following UAST Roles if present on any children:
// If | SwitchCase |  For[Each] | [Do]While | TryCatch | Continue | OpBoolean* | Goto
// Important: since some languages allow for code defined
// outside function definitions, this won't check that the Node has the role FunctionDeclarationRole
// so the user should check that if the intended use is calculating the complexity of a function/method.
// If the children contain more than one function definitions, the value will not be averaged between
// the total number of function declarations but given as a total.
//
// Some practical implementations counting tokens in the code. They sometimes differ; for example
// some of them count the switch "default" as an incrementor, some consider all return values minus the
// last, some of them consider "else" (which is wrong IMHO, but not for elifs, remember than the IfElse
// token in the UAST is really an Else not an "else if", elseifs would have a children If token), some
// consider throw and finally while others only the catch, etc.
//
// Examples:
// PMD reference implementation: http://pmd.sourceforge.net/pmd-4.3.0/xref/net/sourceforge/pmd/rules/CyclomaticComplexity.html
// GMetrics: http://gmetrics.sourceforge.net/gmetrics-CyclomaticComplexityMetric.html
// Go: https://github.com/fzipp/gocyclo/blob/master/gocyclo.go#L214
// SonarQube (include rules for many languages): https://docs.sonarqube.org/display/SONAR/Metrics+-+Complexity
//
// IMPORTANT DISCLAIMER: McCabe definition specifies clearly that boolean operations should increment the
// count in 1 for every boolean element when the language if the language evaluates conditions in
// short-circuit.  Unfortunately in the current version of the UAST we don't specify these language invariants
// and also we still haven't defined how we are going to represent the boolean expressions (which also would
// need a tree transformation process in the pipeline that we lack) so lacking a better way of detecting for
// all  languages what symbols or literals are part of a boolean expression we count the boolean operators
// themselves which should work for short-circuit infix languages but not prefix or infix languages that can
// evaluate more than two items with a single operator.  (FIXME when both things are solved in the UAST
// definition and the SDK).
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
			case If, SwitchCase, For, ForEach, While,
			     DoWhile, TryCatch, Continue, OpBooleanAnd, OpBooleanOr,
			     OpBooleanXor:
				complx++
			}
		}
	}
	return complx
}
