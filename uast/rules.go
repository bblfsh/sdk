package uast

// Rule is a function that takes a node path, executes some code, possibly
// mutating any of the given nodes and returns nil if it was successful or
// an error.
type Rule func(...*Node) error

// Selector is a function that takes a node path and returns true if the node
// should be selected.
type Selector func(...*Node) bool

// Rules takes rules and combines them in a single rule.
func Rules(rules ...Rule) Rule {
	return func(ns ...*Node) error {
		for _, r := range rules {
			if err := r(ns...); err != nil {
				return err
			}
		}

		return nil
	}
}

// Role creates a Rule that sets the given roles in every node matched by the
// selector.
func (s Selector) Role(roles ...Role) Rule {
	return func(ns ...*Node) error {
		if s(ns...) {
			n := ns[len(ns)-1]
			n.Roles = append(n.Roles, roles...)
		}

		return nil
	}
}

// OnPath creates a selector that matches a list of selectors against a path from
// the end.
func OnPath(selectors ...Selector) Selector {
	return func(ns ...*Node) bool {
		stack := selectors

		if len(ns) == 0 || len(stack) == 0 {
			return false
		}

		for i := len(ns) - 1; i >= 0; i-- {
			path := ns[i:]
			selector := stack[len(stack)-1]
			if !selector(path...) {
				continue
			}

			if len(stack) == 1 {
				return true
			}

			stack = stack[:len(stack)-1]
			ns = ns[:i]
		}

		return false
	}
}

// OnNoRole is a selector matching nodes with no role at all.
var OnNoRole Selector = func(ns ...*Node) bool {
	if len(ns) == 0 {
		return false
	}

	n := ns[len(ns)-1]
	return len(n.Roles) == 0
}

// OnInternalType creates a selector that matches one or more internal types.
// If more than one internal type is given, each type will be matched against a
// node in the path starting from the end.
func OnInternalType(path ...string) Selector {
	return func(ns ...*Node) bool {
		if len(path) == 0 {
			return false
		}

		if len(path) > len(ns) {
			return false
		}

		i := len(path) - 1
		j := len(ns) - 1
		for i >= 0 && j >= 0 {
			p := path[i]
			n := ns[j]
			if n.InternalType != p {
				return false
			}

			i--
			j--
		}

		return true
	}
}

// OnInternalRole creates a selector matching internal roles. Matching rules
// are analogous to OnInternalType.
func OnInternalRole(path ...string) Selector {
	return func(ns ...*Node) bool {
		if len(path) == 0 {
			return false
		}

		if len(path) > len(ns) {
			return false
		}

		i := len(path) - 1
		j := len(ns) - 1
		for i >= 0 && j >= 0 {
			p := path[i]
			n := ns[j]
			r, ok := n.Properties[InternalRoleKey]
			if !ok || r != p {
				return false
			}

			i--
			j--
		}

		return true
	}
}
