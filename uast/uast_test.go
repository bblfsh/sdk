package uast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefixTokens(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id3"},
			// Prefix is the default so it doesnt need any role
			{InternalType: "op_prefix", Token: "Prefix+", Children: []*Node{
				{InternalType: "left", Token: "tok_pre_left"},
				{InternalType: "right", Token: "tok_pre_right"},
			}}}}
	result := Tokens(n)
	expected := []string{"id3", "Prefix+", "tok_pre_left", "tok_pre_right"}
	require.Equal(expected, result)
}

func TestPrefixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id3"},
			// Prefix is the default so it doesnt need any role
			{InternalType: "op_prefix", Token: "Prefix+", Children: []*Node{
				{InternalType: "left", Token: "tok_pre_left", Children: []*Node{
					{InternalType: "subleft_1a", Token: "subleft_1a", Children: []*Node{
						{InternalType: "subleft_1a_2a", Token: "subleft_1a_2a"},
						{InternalType: "subleft_1a_2b", Token: "subleft_1a_2b"},
					}},
					{InternalType: "subleft_1b", Token: "subleft_1b", Children: []*Node{
						{InternalType: "subleft_b_2a", Token: "subleft_b_2a"},
						{InternalType: "subleft_b_2b", Token: "subleft_b_2b"},
					}},
				}},
				{InternalType: "right", Token: "tok_pre_right"},
			},
			}}}
	result := Tokens(n)
	expected := []string{"id3", "Prefix+", "tok_pre_left", "subleft_1a", "subleft_1a_2a",
		"subleft_1a_2b", "subleft_1b", "subleft_b_2a", "subleft_b_2b",
		"tok_pre_right"}
	require.Equal(expected, result)
}

func TestPrefixTokensPlain(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id3"},
			// Prefix is the default so it doesnt need any role
			{InternalType: "op_prefix", Token: "Prefix+"},
			{InternalType: "left", Token: "tok_pre_left"},
			{InternalType: "right", Token: "tok_pre_right"},
		}}
	result := Tokens(n)
	expected := []string{"id3", "Prefix+", "tok_pre_left", "tok_pre_right"}
	require.Equal(expected, result)
}

func TestInfixTokens(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id1"},
			{InternalType: "op_infix", Roles: []Role{Infix}, Token: "Infix+", Children: []*Node{
				{InternalType: "left", Token: "tok_in_left"},
				{InternalType: "right", Token: "tok_in_right"},
			}}}}
	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "tok_in_right"}
	require.Equal(expected, result)
}

func TestInfixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id3", Token: "id3"},
			// Prefix is the default so it doesnt need any role
			{InternalType: "op_infix", Token: "op_infix", Roles: []Role{Infix}, Children: []*Node{
				{InternalType: "left", Token: "left", Roles: []Role{Infix}, Children: []*Node{
					{InternalType: "subleft_1a", Token: "subleft_1a", Roles: []Role{Infix}, Children: []*Node{
						{InternalType: "subleft_1a_2a", Token: "subleft_1a_2a"},
						{InternalType: "subleft_1a_2b", Token: "subleft_1a_2b"},
					}},
					{InternalType: "subleft_1b", Token: "subleft_1b", Roles: []Role{Infix}, Children: []*Node{
						{InternalType: "subleft_1b_2a", Token: "subleft_1b_2a"},
						{InternalType: "subleft_1b_2b", Token: "subleft_1b_2b"},
					}},
				}},
				{InternalType: "right", Token: "right"},
			},
			}}}
	result := Tokens(n)
	expected := []string{"id3", "subleft_1a_2a", "subleft_1a", "subleft_1a_2b", "left",
		"subleft_1b_2a", "subleft_1b", "subleft_1b_2b", "op_infix", "right"}

	require.Equal(expected, result)
}

func TestInfixTokensPlain(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id1"},
			{InternalType: "left", Token: "tok_in_left"},
			{InternalType: "op_infix", Roles: []Role{Infix}, Token: "Infix+"},
			{InternalType: "right", Token: "tok_in_right"},
		}}
	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "tok_in_right"}
	require.Equal(expected, result)
}

func TestPostfixTokens(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id2"},
			{InternalType: "op_postfix", Roles: []Role{Postfix}, Token: "Postfix+", Children: []*Node{
				{InternalType: "left", Token: "tok_post_left"},
				{InternalType: "right", Token: "tok_post_right"},
			}}}}
	result := Tokens(n)
	expected := []string{"id2", "tok_post_left", "tok_post_right", "Postfix+"}
	require.Equal(expected, result)
}

func TestPostfixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id2"},
			// Prefix is the default so it doesnt need any role
			{InternalType: "op_postfix", Token: "op_postfix", Roles: []Role{Postfix}, Children: []*Node{
				{InternalType: "left", Token: "left", Roles: []Role{Postfix}, Children: []*Node{
					{InternalType: "subleft_1a", Token: "subleft_1a", Roles: []Role{Postfix}, Children: []*Node{
						{InternalType: "subleft_1a_2a", Token: "subleft_1a_2a"},
						{InternalType: "subleft_1a_2b", Token: "subleft_1a_2b"},
					}},
					{InternalType: "subleft_1b", Token: "subleft_1b", Roles: []Role{Postfix}, Children: []*Node{
						{InternalType: "subleft_1b_2a", Token: "subleft_1b_2a"},
						{InternalType: "subleft_1b_2b", Token: "subleft_1b_2b"},
					}},
				}},
				{InternalType: "right", Token: "right"},
			},
			}}}
	result := Tokens(n)
	expected := []string{"id2", "subleft_1a_2a", "subleft_1a_2b", "subleft_1a", "subleft_1b_2a",
		"subleft_1b_2b", "subleft_1b", "left", "right", "op_postfix"}
	require.Equal(expected, result)
}

func TestPostfixTokensPlain(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id2"},
			{InternalType: "left", Token: "tok_post_left"},
			{InternalType: "right", Token: "tok_post_right"},
			{InternalType: "op_postfix", Roles: []Role{Postfix}, Token: "Postfix+"},
		}}
	result := Tokens(n)
	expected := []string{"id2", "tok_post_left", "tok_post_right", "Postfix+"}
	require.Equal(expected, result)
}

// Test for mixed order roles
func TestOrderTokens(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "id", Token: "id1"},
			{InternalType: "op_infix", Roles: []Role{Infix}, Token: "Infix+", Children: []*Node{
				{InternalType: "left", Token: "tok_in_left"},
				{InternalType: "right", Token: "tok_in_right", Roles: []Role{Postfix}, Children: []*Node{
					{InternalType: "subright1", Token: "subright1"},
					{InternalType: "subright2", Token: "subright2"},
				}},
			}},
			{InternalType: "id", Token: "id2"},
			{InternalType: "op_postfix", Roles: []Role{Postfix}, Token: "Postfix+", Children: []*Node{
				{InternalType: "left", Token: "tok_post_left"},
				// Prefix
				{InternalType: "right", Token: "tok_post_right", Children: []*Node{
					{InternalType: "subright_pre1", Token: "subright_pre1"},
					{InternalType: "subright_pre2", Token: "subright_pre2"},
				}},
			}},
			{InternalType: "id", Token: "id3"},

			// Prefix is the default so it doesnt need any role
			{InternalType: "op_prefix", Token: "Prefix+", Children: []*Node{
				{InternalType: "left", Token: "tok_pre_left"},
				{InternalType: "right", Token: "tok_pre_right", Roles: []Role{Infix}, Children: []*Node{
					{InternalType: "subright_in1", Token: "subright_in1"},
					{InternalType: "subright_in2", Token: "subright_in2"},
				}},
			}}}}

	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "subright1", "subright2", "tok_in_right",
		"id2", "tok_post_left", "tok_post_right", "subright_pre1", "subright_pre2", "Postfix+",
		"id3", "Prefix+", "tok_pre_left", "subright_in1", "tok_pre_right", "subright_in2"}
	require.Equal(expected, result)
}

func TestCyclomaticComplexity(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "module",
		Children: []*Node{
			{InternalType: "root"}, // 1 (initial)
			// Prefix is the default so it doesnt need any role
			{InternalType: "if1", Roles: []Role{If}, Children: []*Node{ // 2 (If)
				{InternalType: "if1else1", Roles: []Role{IfElse}, Children: []*Node{ // 0
					{InternalType: "if1else1foreach", Roles: []Role{ForEach}, Children: []*Node{ // 3 (ForEach)
						{InternalType: "foreach_child1"},                                   // 0
						{InternalType: "foreach_child2_continue", Roles: []Role{Continue}}, // 4 (Continue)
					}},
					{InternalType: "if1else1if", Roles: []Role{If}, Children: []*Node{ // 5 (If)
						{InternalType: "elseif_child1"},                      // 0
						{InternalType: "opAnd", Roles: []Role{OpBooleanAnd}}, // 6 (OpBooleanAnd)
						{InternalType: "elseif_child2"},                      // 0
					}},
				}},
				{InternalType: "break", Roles: []Role{Break}},
			},
			}}}
	require.Equal(CyclomaticComplexity(n), 6)
}

func TestCountChildrenOfRol(t *testing.T) {
	require := require.New(t)

	n1 := &Node{InternalType: "module", Children: []*Node{
		{InternalType: "Statement", Roles: []Role{Statement}},
		{InternalType: "Statement", Roles: []Role{Statement}},
		{InternalType: "If", Roles: []Role{If}},
	}}
	n2 := &Node{InternalType: "module", Children: []*Node{
		{InternalType: "Statement", Roles: []Role{Statement}, Children: []*Node{
			{InternalType: "Statement", Roles: []Role{Statement}, Children: []*Node{
				{InternalType: "If", Roles: []Role{If}},
				{InternalType: "Statemenet", Roles: []Role{Statement}},
			}},
		}},
	}}
	result := n1.CountChildrenOfRole(Statement)
	expect := 2
	require.Equal(expect, result)

	result = n2.CountChildrenOfRole(Statement)
	expect = 1
	require.Equal(expect, result)

	result = n1.DeepCountChildrenOfRole(Statement)
	expect = 2
	require.Equal(expect, result)

	result = n2.DeepCountChildrenOfRole(Statement)
	expect = 3
	require.Equal(expect, result)
}

func TestChildrenOfRole(t *testing.T) {
	require := require.New(t)

	n1 := &Node{InternalType: "module", Children: []*Node{
		{InternalType: "Statement", Roles: []Role{Statement}},
		{InternalType: "Statement", Roles: []Role{Statement}},
		{InternalType: "If", Roles: []Role{If}},
	}}
	n2 := &Node{InternalType: "module", Children: []*Node{
		{InternalType: "Statement", Roles: []Role{Statement}, Children: []*Node{
			{InternalType: "Statement", Roles: []Role{Statement}, Children: []*Node{
				{InternalType: "If", Roles: []Role{If}},
				{InternalType: "Statemenet", Roles: []Role{Statement}},
			}},
		}},
	}}

	result := n1.ChildrenOfRole(Statement)
	expect := 2
	require.Equal(expect, len(result))

	result = n2.ChildrenOfRole(Statement)
	expect = 1
	require.Equal(expect, len(result))

	result = n1.DeepChildrenOfRole(Statement)
	expect = 2
	require.Equal(expect, len(result))

	result = n2.DeepChildrenOfRole(Statement)
	expect = 3
	require.Equal(expect, len(result))
}

func TestContainsRole(t *testing.T) {
	require := require.New(t)
	n := &Node{InternalType: "node", Roles: []Role{Statement, If}}

	result := n.ContainsRole(If)
	require.Equal(true, result)

	result = n.ContainsRole(Switch)
	require.Equal(false, result)
}
