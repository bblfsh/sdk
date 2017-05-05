// Package ann provides a DSL to annotate UAST.
package ann

import (
	"github.com/bblfsh/sdk/uast"
)

type axis int

const (
	self axis = iota
	child
	descendant
	descendantOrSelf
)

// Predicate is a function that takes a node and returns a boolean. It
// is provided as a convenient way of defining a PathMatcher.
type Predicate func(n *uast.Node) bool

// Rule is a conversion rule that can visit a tree, match nodes against
// path matchers and apply actions to the matching node.
type Rule struct {
	axis       axis
	predicates []Predicate
	actions    []Action
	rules      []*Rule
}

// On is the *Rule constructor. It takes a list of predicates and returns a
// new *Rule that matches all of them.
func On(predicates ...Predicate) *Rule {
	return &Rule{predicates: predicates}
}

// Self applies the given rules to nodes matched by the current rule.
func (r *Rule) Self(rules ...*Rule) *Rule {
	return r.addRules(self, rules)
}

// Children applies the given rules to children of nodes matched by the current
// rule.
func (r *Rule) Children(rules ...*Rule) *Rule {
	return r.addRules(child, rules)
}

// Descendants applies the given rules to any descendant matched of nodes matched
// by the current rule.
func (r *Rule) Descendants(rules ...*Rule) *Rule {
	return r.addRules(descendant, rules)
}

// DescendantsOrSelf applies the given rules to self and any descendant matched
// of nodes matched by the current rule.
func (r *Rule) DescendantsOrSelf(rules ...*Rule) *Rule {
	return r.addRules(descendantOrSelf, rules)
}

func (r *Rule) addRules(axis axis, rules []*Rule) *Rule {
	for _, r := range rules {
		r.axis = axis
	}

	r.rules = append(r.rules, rules...)
	return r
}

// Apply applies the rule to the given node.
func (r *Rule) Apply(n *uast.Node) error {
	iter := newMatchPathIter(n, r.axis, r.predicates)
	for {
		p := iter.Next()
		if p.IsEmpty() {
			return nil
		}

		mn := p.Node()
		for _, a := range r.actions {
			if err := a(mn); err != nil {
				return err
			}
		}

		for _, cr := range r.rules {
			if err := cr.Apply(mn); err != nil {
				return err
			}
		}
	}
}

// Roles attaches an action to the rule that adds the given roles.
func (r *Rule) Roles(roles ...uast.Role) *Rule {
	return r.Do(AddRoles(roles...))
}

// RuleError values are returned by the annotation process when a rule
// created by the Error function is activated.  A RuleError wraps the
// desired error and carries the node that provoke the error.
type RuleError interface {
	// Error implements the error interface.
	Error() string
	// Inner returns the wrapped error.
	Inner() error
	// Node returns the offending node.
	Node() *uast.Node
}

type ruleError struct {
	error
	node *uast.Node
}

// implements RuleError.
func (e *ruleError) Inner() error {
	return e.error
}

// implements RuleError.
func (e *ruleError) Node() *uast.Node {
	return e.node
}

// Error makes the rule application fail if the current rule matches.
func (r *Rule) Error(err error) *Rule {
	return r.Do(ReturnError(err))
}

// Do attaches actions to the rule.
func (r *Rule) Do(actions ...Action) *Rule {
	r.actions = append(r.actions, actions...)
	return r
}

// HasInternalType matches a node if its internal type matches the given one.
func HasInternalType(it string) Predicate {
	return func(n *uast.Node) bool {
		if n == nil {
			return false
		}

		return n.InternalType == it
	}
}

// HasProperty matches a node if it has a property matching the given key and value.
func HasProperty(k, v string) Predicate {
	return func(n *uast.Node) bool {
		if n == nil {
			return false
		}

		if n.Properties == nil {
			return false
		}

		prop, ok := n.Properties[k]
		return ok && prop == v
	}
}

// HasInternalRole is a convenience shortcut for:
//
//	HasProperty(uast.InternalRoleKey, r)
//
func HasInternalRole(r string) Predicate {
	return HasProperty(uast.InternalRoleKey, r)
}

// HasChild matches a node that contains a child matching the given predicate.
func HasChild(pred Predicate) Predicate {
	return func(n *uast.Node) bool {
		if n == nil {
			return false
		}

		for _, c := range n.Children {
			if pred(c) {
				return true
			}
		}

		return false
	}
}

// HasToken matches a node if its token matches the given one.
func HasToken(tk string) Predicate {
	return func(n *uast.Node) bool {
		if n == nil {
			return false
		}

		return n.Token == tk
	}
}

// Any matches any path.
var Any Predicate = func(n *uast.Node) bool { return true }

// Not negates a node predicate.
func Not(p Predicate) Predicate {
	return func(n *uast.Node) bool {
		return !p(n)
	}
}

// And returns a predicate that returns true if all the given predicates returns
// true.
func And(ps ...Predicate) Predicate {
	return func(n *uast.Node) bool {
		for _, p := range ps {
			if !p(n) {
				return false
			}
		}

		return true
	}
}

// Or returns a predicate that returns true if any of the given predicates returns
// true.
func Or(ps ...Predicate) Predicate {
	return func(n *uast.Node) bool {
		for _, p := range ps {
			if p(n) {
				return true
			}
		}

		return false
	}
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

// ReturnError creates an action that always returns a RuleError
// wrapping the given error with the offending node information.
func ReturnError(err error) Action {
	return func(n *uast.Node) error {
		return &ruleError{
			error: err,
			node:  n,
		}
	}
}

type matchPathIter struct {
	axis       axis
	predicates []Predicate
	iter       uast.PathStepIter
}

func newMatchPathIter(n *uast.Node, axis axis, predicates []Predicate) uast.PathIter {
	return &matchPathIter{
		axis:       axis,
		predicates: predicates,
		iter:       uast.NewOrderPathIter(uast.NewPath(n)),
	}
}

func (i *matchPathIter) Next() uast.Path {
	for {
		p := i.iter.Next()
		if p.IsEmpty() {
			return p
		}

		switch i.axis {
		case self:
			if len(p) >= len(i.predicates) {
				i.iter.Step()
			}

			if matchPredicates(p, i.predicates) {
				return p
			}
		case child:
			if len(p) > len(i.predicates) {
				i.iter.Step()
			}

			p = p[1:]
			if matchPredicates(p, i.predicates) {
				return p
			}
		case descendant:
			p = p[1:]
			if matchSuffixPredicates(p, i.predicates) {
				return p
			}
		case descendantOrSelf:
			if matchSuffixPredicates(p, i.predicates) {
				return p
			}
		}
	}
}

func matchPredicates(path uast.Path, preds []Predicate) bool {
	if len(path) != len(preds) {
		return false
	}

	for i, pred := range preds {
		if !pred(path[i]) {
			return false
		}
	}

	return true
}

func matchSuffixPredicates(path uast.Path, preds []Predicate) bool {
	if len(path) < len(preds) {
		return false
	}

	j := len(path) - 1
	for i := len(preds) - 1; i >= 0; i-- {
		if !preds[i](path[j]) {
			return false
		}

		j--
	}

	return true
}
