package ann

import (
	"fmt"
	"testing"

	"github.com/bblfsh/sdk/uast"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// go test -v -run 'TestNotationSuite' ./uast/ann

type NotationSuite struct {
	suite.Suite
}

func TestNotationSuite(t *testing.T) {
	suite.Run(t, new(NotationSuite))
}

func (suite *NotationSuite) TestAny() {
	rule := On(Any).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[*] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestNotAny() {
	rule := On(Not(Any)).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[not(*)] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestHasInternalType() {
	rule := On(HasInternalType("foo")).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[@InternalType='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestHasProperty() {
	rule := On(HasProperty("key", "value")).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[@key][@key='value'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestHasInternalRole() {
	rule := On(HasInternalRole("role")).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[@internalRole][@internalRole='role'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestHasChild() {
	rule := On(HasChild(HasInternalType("foo"))).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[child::@InternalType='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestToken() {
	rule := On(HasToken("foo")).Roles(uast.SimpleIdentifier)
	expected := head + "| /self::*[@Token='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestAnd() {
	rule := On(And(
		HasToken("foo"),
		HasToken("bar"),
		HasInternalType("bla"),
	)).Roles(uast.SimpleIdentifier)
	expected := head +
		"| /self::*[(@Token='foo') and (@Token='bar') and (@InternalType='bla')] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestOr() {
	rule := On(Or(
		HasToken("foo"),
		HasToken("bar"),
		HasInternalType("bla"),
	)).Roles(uast.SimpleIdentifier)
	expected := head +
		"| /self::*[(@Token='foo') or (@Token='bar') or (@InternalType='bla')] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestSelf() {
	rule := On(Any).Self(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + "| /self::*[@Token='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)

	rule = On(Any).Self(On(HasToken("foo"))).Roles(uast.SimpleIdentifier)
	expected = head + "| /self::*[*] | SimpleIdentifier |\n"
	obtained = rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestChildren() {
	rule := On(Any).Children(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + "| /*[@Token='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestDescendants() {
	rule := On(Any).Descendants(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + "| //*[@Token='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestDescendantsOrSelf() {
	rule := On(Any).DescendantsOrSelf(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + "| /descendant-or-self::*[@Token='foo'] | SimpleIdentifier |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *NotationSuite) TestMisc1() {
	rule := On(Any).Self(
		On(Not(HasInternalType("FILE"))).Error(fmt.Errorf("root must be CompilationUnit")),
		On(HasInternalType("FILE")).Roles(uast.SimpleIdentifier).Descendants(
			On(HasInternalType("identifier")).Roles(uast.QualifiedIdentifier),
			On(HasInternalType("binary expression")).Children(
				On(HasInternalType("left")).Roles(uast.BinaryExpressionLeft)),
		))
	expected := head +
		"| /self::*[not(@InternalType='FILE')] | Error |\n" +
		"| /self::*[@InternalType='FILE'] | SimpleIdentifier |\n" +
		"| /self::*[@InternalType='FILE']//*[@InternalType='identifier'] | QualifiedIdentifier |\n" +
		"| /self::*[@InternalType='FILE']//*[@InternalType='binary expression']/*[@InternalType='left'] | BinaryExpressionLeft |\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}
