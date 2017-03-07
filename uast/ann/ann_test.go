package ann

import (
	"fmt"
	"testing"

	. "github.com/bblfsh/sdk/uast"

	"github.com/stretchr/testify/require"
)

func TestHasInternalType(t *testing.T) {
	require := require.New(t)

	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.InternalType = s
			path = append(path, n)
		}

		return path
	}

	has := HasInternalType("foo")

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("foo")))
	require.Equal(path("other"), has.MatchPath(path("other")))
	require.Equal(path("other"), has.MatchPath(path("other", "foo")))
	require.Equal(path("foo", "bar"), has.MatchPath(path("foo", "bar")))
}

func TestHasInternalRole(t *testing.T) {
	require := require.New(t)

	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.Properties[InternalRoleKey] = s
			path = append(path, n)
		}

		return path
	}

	has := HasInternalRole("foo")

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("foo")))
	require.Equal(path("other"), has.MatchPath(path("other")))
	require.Equal(path("other"), has.MatchPath(path("other", "foo")))
	require.Equal(path("foo", "bar"), has.MatchPath(path("foo", "bar")))
}

func TestHasProperty(t *testing.T) {
	require := require.New(t)

	myProp := "property"
	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.Properties[myProp] = s
			path = append(path, n)
		}

		return path
	}

	has := HasProperty(myProp, "foo")

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("foo")))
	require.Equal(path("other"), has.MatchPath(path("other")))
	require.Equal(path("other"), has.MatchPath(path("other", "foo")))
	require.Equal(path("foo", "bar"), has.MatchPath(path("foo", "bar")))
}

func TestHasToken(t *testing.T) {
	require := require.New(t)

	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.Token = s
			path = append(path, n)
		}

		return path
	}

	has := HasToken("foo")

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("foo")))
	require.Equal(path(""), has.MatchPath(path("")))
	require.Equal(path("other"), has.MatchPath(path("other")))
	require.Equal(path("other"), has.MatchPath(path("other", "foo")))
	require.Equal(path("foo", "bar"), has.MatchPath(path("foo", "bar")))

	has = HasToken("")

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("")))
	require.Equal(path("other"), has.MatchPath(path("other")))
	require.Equal(path("other"), has.MatchPath(path("other", "")))
	require.Equal(path("", "bar"), has.MatchPath(path("", "bar")))
}

func TestAny(t *testing.T) {
	require := require.New(t)

	myProp := "property"
	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.Properties[myProp] = s
			path = append(path, n)
		}

		return path
	}

	has := Any()

	require.Equal(path(), has.MatchPath(path()))
	require.Equal(path(), has.MatchPath(path("foo")))
	require.Equal(path(), has.MatchPath(path("other")))
	require.Equal(path(), has.MatchPath(path("other", "foo")))
	require.Equal(path(), has.MatchPath(path("foo", "bar")))
}

func TestNot(t *testing.T) {
	require := require.New(t)

	path := func(ss ...string) Path {
		path := NewPath()
		for _, s := range ss {
			n := NewNode()
			n.InternalType = s
			path = append(path, n)
		}

		return path
	}

	has := Not(HasInternalType("foo"))

	require.Equal(path("foo"), has.MatchPath(path("foo")))
	require.Equal(path(), has.MatchPath(path("other")))
	require.Equal(path("other", "foo"), has.MatchPath(path("other", "foo")))
	require.Equal(path("foo"), has.MatchPath(path("foo", "bar")))
}

func TestAddRoles(t *testing.T) {
	require := require.New(t)

	a := AddRoles(Statement, Expression)
	input := NewNode()
	expected := NewNode()
	expected.Roles = []Role{Statement, Expression}
	err := a(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any()).Roles(role)

	input := &Node{
		InternalType: "root",
		Children: []*Node{{
			InternalType: "foo",
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{role},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{role},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnRulesApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(HasInternalType("root")).
		Rules(On(HasInternalType("foo")).Roles(role))

	input := &Node{
		InternalType: "root",
		Children: []*Node{{
			InternalType: "foo",
		}},
	}
	expected := &Node{
		InternalType: "root",
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{role},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnRulesActionError(t *testing.T) {
	require := require.New(t)

	errorAction := func(n *Node) error {
		return fmt.Errorf("test error")
	}

	rule := On(HasInternalType("root")).
		Rules(On(HasInternalType("foo")).Do(errorAction))

	input := &Node{
		InternalType: "root",
		Children: []*Node{{
			InternalType: "foo",
		}},
	}
	err := rule.Apply(input)
	require.EqualError(err, "test error")
}

func TestPathPredicateMatchPath(t *testing.T) {
	require := require.New(t)

	falsePred := PathPredicate(func(path Path) bool {
		return false
	})

	unmatched := falsePred.MatchPath(NewPath(&Node{}))
	require.Len(unmatched, 1)
}
