package ann

import (
	"testing"

	errors "gopkg.in/src-d/go-errors.v0"

	. "gopkg.in/bblfsh/sdk.v1/uast"

	"github.com/stretchr/testify/require"
)

func TestHasInternalType(t *testing.T) {
	require := require.New(t)

	node := func(s string) *Node {
		n := NewNode()
		n.InternalType = s
		return n
	}

	pred := HasInternalType("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))
}

func TestHasInternalRole(t *testing.T) {
	require := require.New(t)

	node := func(s string) *Node {
		n := NewNode()
		n.Properties[InternalRoleKey] = s
		return n
	}

	pred := HasInternalRole("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))
}

func TestHasProperty(t *testing.T) {
	require := require.New(t)

	node := func(k, v string) *Node {
		n := NewNode()
		n.Properties[k] = v
		return n
	}

	pred := HasProperty("myprop", "foo")
	require.True(pred.Eval(node("myprop", "foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("myprop", "bar")))
	require.False(pred.Eval(node("otherprop", "foo")))
	require.False(pred.Eval(node("otherprop", "bar")))
}

func TestHasToken(t *testing.T) {
	require := require.New(t)

	node := func(s string) *Node {
		n := NewNode()
		n.Token = s
		return n
	}

	pred := HasToken("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))

	pred = HasToken("")
	require.True(pred.Eval(&Node{}))
	require.True(pred.Eval(node("")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(node("bar")))
}

func TestAny(t *testing.T) {
	require := require.New(t)

	pred := Any
	require.True(pred.Eval(nil))
	require.True(pred.Eval(&Node{}))
	require.True(pred.Eval(NewNode()))
	require.True(pred.Eval(&Node{InternalType: "foo"}))
	require.True(pred.Eval(&Node{Token: "foo"}))
}

func TestNot(t *testing.T) {
	require := require.New(t)

	node := func(s string) *Node {
		n := NewNode()
		n.InternalType = s
		return n
	}

	pred := Not(HasInternalType("foo"))
	require.False(pred.Eval(node("foo")))
	require.True(pred.Eval(nil))
	require.True(pred.Eval(&Node{}))
	require.True(pred.Eval(node("")))
	require.True(pred.Eval(node("bar")))
}

func TestOr(t *testing.T) {
	require := require.New(t)

	node := func(s string) *Node {
		n := NewNode()
		n.InternalType = s
		return n
	}

	pred := Or(HasInternalType("foo"), HasInternalType("bar"))
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("")))
	require.True(pred.Eval(node("bar")))
	require.False(pred.Eval(node("baz")))
}

func TestAnd(t *testing.T) {
	require := require.New(t)

	node := func(typ, tok string) *Node {
		n := NewNode()
		n.InternalType = typ
		n.Token = tok
		return n
	}

	pred := And(HasInternalType("foo"), HasToken("bar"))
	require.False(pred.Eval(node("foo", "")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.False(pred.Eval(node("", "")))
	require.False(pred.Eval(node("bar", "")))
	require.False(pred.Eval(node("foo", "foo")))
	require.False(pred.Eval(node("bar", "bar")))
	require.True(pred.Eval(node("foo", "bar")))
}

func TestHasChild(t *testing.T) {
	require := require.New(t)

	pred := HasChild(HasInternalType("foo"))

	path := func(s ...string) *Node {
		var n *Node
		for i := len(s) - 1; i >= 0; i-- {
			tn := NewNode()
			tn.InternalType = s[i]
			if n != nil {
				tn.Children = append(tn.Children, n)
			}

			n = tn
		}

		return n
	}

	require.False(pred.Eval(path("foo")))
	require.False(pred.Eval(path("foo", "bar")))
	require.False(pred.Eval(path("", "")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(&Node{}))
	require.True(pred.Eval(path("bar", "foo")))
	require.False(pred.Eval(path("bar", "baz", "foo")))
}

func TestAddRoles(t *testing.T) {
	require := require.New(t)

	a := AddRoles(Statement, Expression)
	input := NewNode()
	expected := NewNode()
	expected.Roles = []Role{Statement, Expression}
	err := a.Do(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any).Roles(role)

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{role},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnSelfApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any).Self(On(HasInternalType("root")).Roles(role))

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "bar",
				Roles:        []Role{Unannotated},
				Children: []*Node{{
					InternalType: "baz",
					Roles:        []Role{Unannotated},
				}},
			}},
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{role},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "bar",
				Roles:        []Role{Unannotated},
				Children: []*Node{{
					InternalType: "baz",
					Roles:        []Role{Unannotated},
				}},
			}},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnChildrenApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any).Children(On(HasInternalType("foo")).Roles(role))

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{role},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
		}},
	}
	expected = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{Unannotated},
			}},
		}},
	}
	expected = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{Unannotated},
			}},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnDescendantsApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any).Descendants(On(HasInternalType("foo")).Roles(role))

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{role},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
		}},
	}
	expected = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{Unannotated},
			}},
		}},
	}
	expected = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{role},
			}},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnDescendantsOrSelfApply(t *testing.T) {
	require := require.New(t)

	role := Block
	rule := On(Any).DescendantsOrSelf(On(HasInternalType("foo")).Roles(role))

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	expected := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{role},
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = &Node{
		InternalType: "foo",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{Unannotated},
			}},
		}},
	}
	expected = &Node{
		InternalType: "foo",
		Roles:        []Role{role},
		Children: []*Node{{
			InternalType: "bar",
			Roles:        []Role{Unannotated},
			Children: []*Node{{
				InternalType: "foo",
				Roles:        []Role{role},
			}},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnRulesActionError(t *testing.T) {
	require := require.New(t)

	var ErrTestMe = errors.NewKind("test me: %s")
	rule := On(HasInternalType("root")).
		Children(On(HasInternalType("foo")).Error(ErrTestMe.New("foo node found")))

	input := &Node{
		InternalType: "root",
		Roles:        []Role{Unannotated},
		Children: []*Node{{
			InternalType: "foo",
			Roles:        []Role{Unannotated},
		}},
	}
	err := rule.Apply(input)
	require.EqualError(err, "test me: foo node found")

	extraInfoError, ok := err.(RuleError)
	require.Equal(ok, true)
	require.EqualError(extraInfoError, "test me: foo node found")
	require.True(ErrTestMe.Is(extraInfoError.Inner()))

	offendingNode := extraInfoError.Node()
	require.Equal(offendingNode.InternalType, "foo")
}
