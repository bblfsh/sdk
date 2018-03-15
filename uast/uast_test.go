package uast

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v1/uast/role"
)

func tObj(typ, tok string) Object {
	obj := Object{KeyType: String(typ)}
	if tok != "" {
		obj[KeyToken] = String(tok)
	}
	return obj
}

func TestPrefixTokens(t *testing.T) {
	require := require.New(t)

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id3"),
			// Prefix is the default so it doesnt need any role
			Object{
				KeyType: String("op_prefix"), KeyToken: String("Prefix+"),
				"b": List{
					tObj("left", "tok_pre_left"),
					tObj("right", "tok_pre_right"),
				},
			}}}
	result := Tokens(n)
	expected := []string{"id3", "Prefix+", "tok_pre_left", "tok_pre_right"}
	require.Equal(expected, result)
}

func TestPrefixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id3"),
			// Prefix is the default so it doesnt need any role
			Object{KeyType: String("op_prefix"), KeyToken: String("Prefix+"), "b": List{
				Object{KeyType: String("left"), KeyToken: String("tok_pre_left"), "c": List{
					Object{KeyType: String("subleft_1a"), KeyToken: String("subleft_1a"), "d": List{
						tObj("subleft_1a_2a", "subleft_1a_2a"),
						tObj("subleft_1a_2b", "subleft_1a_2b"),
					}},
					Object{KeyType: String("subleft_1b"), KeyToken: String("subleft_1b"), "e": List{
						tObj("subleft_b_2a", "subleft_b_2a"),
						tObj("subleft_b_2b", "subleft_b_2b"),
					}},
				}},
				tObj("right", "tok_pre_right"),
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

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id3"),
			// Prefix is the default so it doesnt need any role
			tObj("op_prefix", "Prefix+"),
			tObj("left", "tok_pre_left"),
			tObj("right", "tok_pre_right"),
		}}
	result := Tokens(n)
	expected := []string{"id3", "Prefix+", "tok_pre_left", "tok_pre_right"}
	require.Equal(expected, result)
}

func TestInfixTokens(t *testing.T) {
	require := require.New(t)
	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id1"),
			Object{KeyType: String("op_infix"), KeyToken: String("Infix+"), KeyRoles: RoleList(role.Infix), "b": List{
				tObj("left", "tok_in_left"),
				tObj("right", "tok_in_right"),
			}}}}
	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "tok_in_right"}
	require.Equal(expected, result)
}

func TestInfixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id3", "id3"),
			// Prefix is the default so it doesnt need any role
			Object{KeyType: String("op_infix"), KeyToken: String("op_infix"), KeyRoles: RoleList(role.Infix), "b": List{
				Object{KeyType: String("left"), KeyToken: String("left"), KeyRoles: RoleList(role.Infix), "c": List{
					Object{KeyType: String("subleft_1a"), KeyToken: String("subleft_1a"), KeyRoles: RoleList(role.Infix), "d": List{
						tObj("subleft_1a_2a", "subleft_1a_2a"),
						tObj("subleft_1a_2b", "subleft_1a_2b"),
					}},
					Object{KeyType: String("subleft_1b"), KeyToken: String("subleft_1b"), KeyRoles: RoleList(role.Infix), "e": List{
						tObj("subleft_1b_2a", "subleft_1b_2a"),
						tObj("subleft_1b_2b", "subleft_1b_2b"),
					}},
				}},
				tObj("right", "right"),
			},
			}}}
	result := Tokens(n)
	expected := []string{"id3", "subleft_1a_2a", "subleft_1a", "subleft_1a_2b", "left",
		"subleft_1b_2a", "subleft_1b", "subleft_1b_2b", "op_infix", "right"}

	require.Equal(expected, result)
}

func TestInfixTokensPlain(t *testing.T) {
	require := require.New(t)
	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id1"),
			tObj("left", "tok_in_left"),
			Object{KeyType: String("op_infix"), KeyToken: String("Infix+"), KeyRoles: RoleList(role.Infix)},
			tObj("right", "tok_in_right"),
		}}
	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "tok_in_right"}
	require.Equal(expected, result)
}

func TestPostfixTokens(t *testing.T) {
	require := require.New(t)
	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id2"),
			Object{KeyType: String("op_postfix"), KeyToken: String("Postfix+"), KeyRoles: RoleList(role.Postfix), "b": List{
				tObj("left", "tok_post_left"),
				tObj("right", "tok_post_right"),
			}}}}
	result := Tokens(n)
	expected := []string{"id2", "tok_post_left", "tok_post_right", "Postfix+"}
	require.Equal(expected, result)
}

func TestPostfixTokensSubtree(t *testing.T) {
	require := require.New(t)

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id2"),
			// Prefix is the default so it doesnt need any role
			Object{KeyType: String("op_postfix"), KeyToken: String("op_postfix"), KeyRoles: RoleList(role.Postfix), "b": List{
				Object{KeyType: String("left"), KeyToken: String("left"), KeyRoles: RoleList(role.Postfix), "c": List{
					Object{KeyType: String("subleft_1a"), KeyToken: String("subleft_1a"), KeyRoles: RoleList(role.Postfix), "d": List{
						tObj("subleft_1a_2a", "subleft_1a_2a"),
						tObj("subleft_1a_2b", "subleft_1a_2b"),
					}},
					Object{KeyType: String("subleft_1b"), KeyToken: String("subleft_1b"), KeyRoles: RoleList(role.Postfix), "e": List{
						tObj("subleft_1b_2a", "subleft_1b_2a"),
						tObj("subleft_1b_2b", "subleft_1b_2b"),
					}},
				}},
				tObj("right", "right"),
			},
			}}}
	result := Tokens(n)
	expected := []string{"id2", "subleft_1a_2a", "subleft_1a_2b", "subleft_1a", "subleft_1b_2a",
		"subleft_1b_2b", "subleft_1b", "left", "right", "op_postfix"}
	require.Equal(expected, result)
}

func TestPostfixTokensPlain(t *testing.T) {
	require := require.New(t)
	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id2"),
			tObj("left", "tok_post_left"),
			tObj("right", "tok_post_right"),
			Object{KeyType: String("op_postfix"), KeyToken: String("Postfix+"), KeyRoles: RoleList(role.Postfix)},
		}}
	result := Tokens(n)
	expected := []string{"id2", "tok_post_left", "tok_post_right", "Postfix+"}
	require.Equal(expected, result)
}

// Test for mixed order roles
func TestOrderTokens(t *testing.T) {
	require := require.New(t)

	n := Object{KeyType: String("module"),
		"a": List{
			tObj("id", "id1"),
			Object{KeyType: String("op_infix"), KeyToken: String("Infix+"), KeyRoles: RoleList(role.Infix), "b": List{
				tObj("left", "tok_in_left"),
				Object{KeyType: String("right"), KeyToken: String("tok_in_right"), KeyRoles: RoleList(role.Postfix), "c": List{
					tObj("subright1", "subright1"),
					tObj("subright2", "subright2"),
				}},
			}},
			tObj("id", "id2"),
			Object{KeyType: String("op_postfix"), KeyToken: String("Postfix+"), KeyRoles: RoleList(role.Postfix), "d": List{
				tObj("left", "tok_post_left"),
				// Prefix
				Object{KeyType: String("right"), KeyToken: String("tok_post_right"), "e": List{
					tObj("subright_pre1", "subright_pre1"),
					tObj("subright_pre2", "subright_pre2"),
				}},
			}},
			tObj("id", "id3"),

			// Prefix is the default so it doesnt need any role
			Object{KeyType: String("op_prefix"), KeyToken: String("Prefix+"), "f": List{
				tObj("left", "tok_pre_left"),
				Object{KeyType: String("right"), KeyToken: String("tok_pre_right"), KeyRoles: RoleList(role.Infix), "g": List{
					tObj("subright_in1", "subright_in1"),
					tObj("subright_in2", "subright_in2"),
				}},
			}}}}

	result := Tokens(n)
	expected := []string{"id1", "tok_in_left", "Infix+", "subright1", "subright2", "tok_in_right",
		"id2", "tok_post_left", "tok_post_right", "subright_pre1", "subright_pre2", "Postfix+",
		"id3", "Prefix+", "tok_pre_left", "subright_in1", "tok_pre_right", "subright_in2"}
	require.Equal(expected, result)
}
