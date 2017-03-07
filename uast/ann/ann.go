// Package ann provides a DSL to annotate UAST.
package ann

import (
	"github.com/bblfsh/sdk/uast"
)

// PathMatcher is a function that matches against a NodePath. If a prefix is not
// matched, it returns that prefix. If there is no match at all, it returns the
// input NodePath. If it matches the full NodePath, nil is returned.
type PathMatcher interface {
	// MatchPath matches against the given path and returns any unmatched
	// prefix.
	MatchPath(path uast.NodePath) uast.NodePath
}

// PathPredicate is a function that takes a node path and returns a boolean. It
// is provided as a convenient way of defining a PathMatcher.
type PathPredicate func(path uast.NodePath) bool

// MatchPath returns an empty path if the PathPredicate is true. Otherwise, it
// returns the given path.
func (p PathPredicate) MatchPath(path uast.NodePath) uast.NodePath {
	if p(path) {
		return uast.NewNodePath()
	}

	return path
}

// NodePredicate is a function that takes a node and returns a boolean. It
// is provided as a convenient way of defining a PathMatcher.
type NodePredicate func(n *uast.Node) bool

// MatchPath returns the parent of the given path if the NodePredicate is true.
// Otherwise, it returns the given path.
func (p NodePredicate) MatchPath(path uast.NodePath) uast.NodePath {
	if path.IsEmpty() {
		return path
	}

	n := path.Node()
	if !p(n) {
		return path
	}

	return path.Parent()
}

// Rule is a conversion rule that can visit a tree, match nodes against
// path matchers and apply actions to the matching node.
type Rule struct {
	on      PathMatcher
	actions []Action
	rules   []*Rule
}

// On is the *Rule constructor. It takes a list of path matchers and returns a
// new *Rule with all the given matchers joined (see the `Join` function).
func On(matchers ...PathMatcher) *Rule {
	return &Rule{
		on: Join(matchers...),
	}
}

// Apply applies the rule to the given node.
func (r *Rule) Apply(n *uast.Node) error {
	var ruleStack [][]*Rule
	ruleStack = append(ruleStack, []*Rule{r})
	pathStack := []uast.NodePath{uast.NewNodePath(n)}

	for {
		lvl := len(pathStack)
		if lvl == 0 {
			break
		}

		path := pathStack[lvl-1]
		pathStack = pathStack[:lvl-1]
		rules := ruleStack[lvl-1]
		ruleStack = ruleStack[:lvl-1]
		var childRules []*Rule
		childRules = append(childRules, rules...)
		n := path.Node()

		for _, r := range rules {
			childRules = append(childRules, r.rules...)
			if len(r.on.MatchPath(path)) > 0 {
				continue
			}

			for _, a := range r.actions {
				if err := a(n); err != nil {
					return err
				}
			}
		}

		for _, child := range n.Children {
			pathStack = append(pathStack, append(path, child))
			ruleStack = append(ruleStack, childRules)
		}
	}

	return nil
}

// Rules attaches a list of rules as children of the current rule.
func (r *Rule) Rules(rules ...*Rule) *Rule {
	for _, or := range rules {
		or.on = Join(r.on, or.on)
	}

	r.rules = append(r.rules, rules...)
	return r
}

// Roles attaches an action to the rule that adds the given roles.
func (r *Rule) Roles(roles ...uast.Role) *Rule {
	return r.Do(AddRoles(roles...))
}

// Do attaches actions to the rule.
func (r *Rule) Do(actions ...Action) *Rule {
	r.actions = append(r.actions, actions...)
	return r
}

// HasInternalType matches a node if its internal type matches the given one.
func HasInternalType(it string) NodePredicate {
	return func(n *uast.Node) bool {
		return n.InternalType == it
	}
}

// HasProperty matches a node if it has a property matching the given key and value.
func HasProperty(k, v string) NodePredicate {
	return func(n *uast.Node) bool {
		prop, ok := n.Properties[k]
		return ok && prop == v
	}
}

// HasInternalRole is a convenience shortcut for:
//
//	HasProperty(uast.InternalRoleKey, r)
//
func HasInternalRole(r string) NodePredicate {
	return HasProperty(uast.InternalRoleKey, r)
}

// HasToken matches a node if its token matches the given one.
func HasToken(tk string) NodePredicate {
	return func(n *uast.Node) bool {
		return n.Token == tk
	}
}

// Any matches any path.
func Any() PathPredicate {
	return func(uast.NodePath) bool { return true }
}

// Not negates a node predicate.
func Not(p NodePredicate) NodePredicate {
	return func(n *uast.Node) bool {
		return !p(n)
	}
}

type joinMatcher struct {
	matchers []PathMatcher
}

// Join joins the given path matchers into a single one.
func Join(matchers ...PathMatcher) PathMatcher {
	jm := &joinMatcher{}
	for _, m := range matchers {
		if ojm, ok := m.(*joinMatcher); ok {
			jm.matchers = append(jm.matchers, ojm.matchers...)
		} else {
			jm.matchers = append(jm.matchers, m)
		}
	}

	return jm
}

func (m *joinMatcher) MatchPath(path uast.NodePath) uast.NodePath {
	stack := m.matchers
	for {
		if len(stack) == 0 {
			break
		}

		matcher := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		n := len(path)
		path = matcher.MatchPath(path)
		if len(path) == n {
			break
		}
	}

	return path
}

// Action is a function that takes a node, does something with it
// (possibly mutating it) and returns an optional error.
type Action func(n *uast.Node) error

// AddRoles creates an action to add the given roles to a node.
func AddRoles(roles ...uast.Role) Action {
	return func(n *uast.Node) error {
		n.Roles = append(n.Roles, roles...)
		return nil
	}
}
