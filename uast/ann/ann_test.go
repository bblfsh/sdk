package ann

import (
	"testing"

	"gopkg.in/bblfsh/sdk.v1/uast"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
	"gopkg.in/src-d/go-errors.v1"
)

func TestHasInternalType(t *testing.T) {
	require := require.New(t)

	node := func(s string) uast.Node {
		n := uast.NewNode()
		n.SetType(s)
		return n
	}

	pred := HasInternalType("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))
}

/*
func TestHasInternalRole(t *testing.T) {
	require := require.New(t)

	node := func(s string) uast.Node {
		n := uast.NewNode()
		n.Properties[uast.InternalRoleKey] = s
		return n
	}

	pred := HasInternalRole("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))
}
*/
func TestHasProperty(t *testing.T) {
	require := require.New(t)

	node := func(k, v string) uast.Node {
		n := uast.NewNode()
		n.SetProperty(k, v)
		return n
	}

	pred := HasProperty("myprop", "foo")
	require.True(pred.Eval(node("myprop", "foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("myprop", "bar")))
	require.False(pred.Eval(node("otherprop", "foo")))
	require.False(pred.Eval(node("otherprop", "bar")))
}

func TestHasToken(t *testing.T) {
	require := require.New(t)

	node := func(s string) uast.Node {
		n := uast.NewNode()
		n.SetToken(s)
		return n
	}

	pred := HasToken("foo")
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("")))
	require.False(pred.Eval(node("bar")))

	pred = HasToken("")
	require.True(pred.Eval(uast.EmptyNode()))
	require.True(pred.Eval(node("")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(node("bar")))
}

func TestAny(t *testing.T) {
	require := require.New(t)

	pred := Any
	require.True(pred.Eval(nil))
	require.True(pred.Eval(uast.EmptyNode()))
	require.True(pred.Eval(uast.NewNode()))
	require.True(pred.Eval(uast.EmptyNode().SetType("foo")))
	require.True(pred.Eval(uast.EmptyNode().SetToken("foo")))
}

func TestNot(t *testing.T) {
	require := require.New(t)

	node := func(s string) uast.Node {
		n := uast.NewNode()
		n.SetType(s)
		return n
	}

	pred := Not(HasInternalType("foo"))
	require.False(pred.Eval(node("foo")))
	require.True(pred.Eval(nil))
	require.True(pred.Eval(uast.EmptyNode()))
	require.True(pred.Eval(node("")))
	require.True(pred.Eval(node("bar")))
}

func TestOr(t *testing.T) {
	require := require.New(t)

	node := func(s string) uast.Node {
		n := uast.NewNode()
		n.SetType(s)
		return n
	}

	pred := Or(HasInternalType("foo"), HasInternalType("bar"))
	require.True(pred.Eval(node("foo")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("")))
	require.True(pred.Eval(node("bar")))
	require.False(pred.Eval(node("baz")))
}

func TestAnd(t *testing.T) {
	require := require.New(t)

	node := func(typ, tok string) uast.Node {
		n := uast.NewNode()
		n.SetType(typ)
		n.SetToken(tok)
		return n
	}

	pred := And(HasInternalType("foo"), HasToken("bar"))
	require.False(pred.Eval(node("foo", "")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.False(pred.Eval(node("", "")))
	require.False(pred.Eval(node("bar", "")))
	require.False(pred.Eval(node("foo", "foo")))
	require.False(pred.Eval(node("bar", "bar")))
	require.True(pred.Eval(node("foo", "bar")))
}

func TestHasChild(t *testing.T) {
	require := require.New(t)

	pred := HasChild(HasInternalType("foo"))

	path := func(s ...string) uast.Node {
		var n uast.Node
		for i := len(s) - 1; i >= 0; i-- {
			tn := uast.NewNode()
			tn.SetType(s[i])
			if n != nil {
				tn["child"] = n
			}

			n = tn
		}

		return n
	}

	require.False(pred.Eval(path("foo")))
	require.False(pred.Eval(path("foo", "bar")))
	require.False(pred.Eval(path("", "")))
	require.False(pred.Eval(nil))
	require.False(pred.Eval(uast.EmptyNode()))
	require.True(pred.Eval(path("bar", "foo")))
	require.False(pred.Eval(path("bar", "baz", "foo")))
}

func TestAddRoles(t *testing.T) {
	require := require.New(t)

	a := AddRoles(role.Statement, role.Expression)
	input := uast.NewNode()
	expected := uast.NewNode()
	expected.SetRoles(role.Statement, role.Expression)
	err := a.Do(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestAddDuplicatedRoles(t *testing.T) {
	require := require.New(t)

	a := AddRoles(role.Statement, role.Expression, role.Statement, role.Expression,
		role.Call, role.Call)
	input := uast.NewNode()
	expected := uast.NewNode()
	expected.SetRoles(role.Statement, role.Expression, role.Call)
	err := a.Do(input)
	require.NoError(err)
	require.Equal(expected, input)
	err = a.Do(input)
	require.Equal(expected, input)
}

func TestRuleOnApply(t *testing.T) {
	require := require.New(t)

	r := role.Block
	rule := On(Any).Roles(r)

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(r),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnSelfApply(t *testing.T) {
	require := require.New(t)

	r := role.Block
	rule := On(Any).Self(On(HasInternalType("root")).Roles(r))

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("bar"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
				"children": uast.List{uast.Object{
					uast.KeyType:  uast.String("baz"),
					uast.KeyRoles: uast.RoleList(role.Unannotated),
				}},
			}},
		}},
	}
	expected := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(r),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("bar"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
				"children": uast.List{uast.Object{
					uast.KeyType:  uast.String("baz"),
					uast.KeyRoles: uast.RoleList(role.Unannotated),
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

	r := role.Block
	rule := On(Any).Children(On(HasInternalType("foo")).Roles(r))

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(r),
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
			}},
		}},
	}
	expected = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
			}},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnDescendantsApply(t *testing.T) {
	require := require.New(t)

	r := role.Block
	rule := On(Any).Descendants(On(HasInternalType("foo")).Roles(r))

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(r),
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
			}},
		}},
	}
	expected = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(r),
			}},
		}},
	}
	err = rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)
}

func TestRuleOnDescendantsOrSelfApply(t *testing.T) {
	require := require.New(t)

	r := role.Block
	rule := On(Any).DescendantsOrSelf(On(HasInternalType("foo")).Roles(r))

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	expected := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(r),
		}},
	}
	err := rule.Apply(input)
	require.NoError(err)
	require.Equal(expected, input)

	input = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(role.Unannotated),
			}},
		}},
	}
	expected = uast.Object{
		uast.KeyType:  uast.String("foo"),
		uast.KeyRoles: uast.RoleList(r),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("bar"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
			"children": uast.List{uast.Object{
				uast.KeyType:  uast.String("foo"),
				uast.KeyRoles: uast.RoleList(r),
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

	input := uast.Object{
		uast.KeyType:  uast.String("root"),
		uast.KeyRoles: uast.RoleList(role.Unannotated),
		"children": uast.List{uast.Object{
			uast.KeyType:  uast.String("foo"),
			uast.KeyRoles: uast.RoleList(role.Unannotated),
		}},
	}
	err := rule.Apply(input)
	require.EqualError(err, "test me: foo node found")

	extraInfoError, ok := err.(RuleError)
	require.Equal(ok, true)
	require.EqualError(extraInfoError, "test me: foo node found")
	require.True(ErrTestMe.Is(extraInfoError.Inner()))

	offendingNode, _ := extraInfoError.Node().(uast.Object)
	require.Equal(offendingNode.Type(), "foo")
}

func TestBetterErrorMessageForInorderTraversalOfNonBinaryNode(t *testing.T) {
	require := require.New(t)

	rule := On(Any).DescendantsOrSelf(
		On(HasInternalType("foo")).
			Roles(role.Infix).
			Children(On(Any).Roles(role.Call)),
	)

	input := uast.Object{
		uast.KeyType: uast.String("foo"),
		"children": uast.List{
			uast.Object{uast.KeyType: uast.String("child")},
			uast.Object{uast.KeyType: uast.String("child")},
			uast.Object{uast.KeyType: uast.String("child")},
		},
	}

	err := rule.Apply(input)
	require.EqualError(err, "unsupported iteration over node with 3 children")
}
