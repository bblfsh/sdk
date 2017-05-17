package ann

import (
	"bytes"
	"fmt"
	"strings"
)

// folder as in "analyzer of recursive data structures combining the
// information in its nodes", not as in "directory".  This particular
// folder traverses Rules in pre-order, generating their string
// XPath-like descriptions.
type folder struct {
	current      path
	descriptions []string
}

// A path represents how to get from the root of a rule to one of its
// nodes.  We push nodes as we go deeper into the tree and pop them back
// when we climb towards its root.
type path []*Rule

func (p *path) push(rule *Rule) {
	*p = append(*p, rule)
}

func (p *path) pop() {
	(*p)[len(*p)-1] = nil
	*p = (*p)[:len(*p)-1]
}

// Returns the path in a format similar to XPath.
func (p *path) String() string {
	var buf bytes.Buffer
	for _, r := range *p {
		buf.WriteRune('/')
		fmt.Fprintf(&buf, "%s::*", r.axis)
		for _, p := range r.predicates {
			fmt.Fprintf(&buf, "[%s]", p)
		}
	}
	return buf.String()
}

// Calculates the description for all the nodes in the rule.
func (f *folder) fold(r *Rule) {
	if len(r.actions) == 0 && len(r.rules) == 0 {
		return
	}

	(&f.current).push(r)
	defer f.current.pop()

	if len(r.actions) != 0 {
		s := fmt.Sprintf("%s -> %s",
			f.current.String(), joinActions(r.actions, ", "))
		f.descriptions = append(f.descriptions, abbreviate(s))
	}

	for _, child := range r.rules {
		f.fold(child)
	}
}

func joinActions(as []Action, sep string) string {
	var buf bytes.Buffer
	_sep := ""
	for _, e := range as {
		fmt.Fprintf(&buf, "%s%s", _sep, e)
		_sep = sep
	}
	return buf.String()
}

// Idempotent.
func abbreviate(s string) string {
	// Replace the On(Any).Something at the begining with root
	if !strings.HasPrefix(s, "/self::*[*] -> ") {
		s = strings.TrimPrefix(s, "/self::*[*]")
	}
	// replace descendant:: with //
	s = strings.Replace(s, "/descendant::*", "//*", -1) // no limit
	// replace child:: with /
	s = strings.Replace(s, "/child::", "/", -1) // no limit
	return s
}

// Returns a string with all the description separated by a newline.
func (f *folder) String() string {
	var buf bytes.Buffer
	for _, e := range f.descriptions {
		fmt.Fprintf(&buf, "%s\n", e)
	}
	return buf.String()
}
