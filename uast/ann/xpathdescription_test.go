package ann

import (
	"fmt"
	"testing"

	"github.com/bblfsh/sdk/uast"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// go test -v -run 'TestRulesDocSuite' ./uast/ann

type RulesDocSuite struct {
	suite.Suite
}

func TestRulesDocSuite(t *testing.T) {
	suite.Run(t, new(RulesDocSuite))
}

func (suite *RulesDocSuite) TestAny() {
	rule := On(Any).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[\*\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestNotAny() {
	rule := On(Not(Any)).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[not\(\*\)\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestHasInternalType() {
	rule := On(HasInternalType("foo")).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[@InternalType='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestHasProperty() {
	rule := On(HasProperty("key", "value")).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[@key\]\[@key='value'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestHasInternalRole() {
	rule := On(HasInternalRole("role")).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[@internalRole\]\[@internalRole='role'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestHasChild() {
	rule := On(HasChild(HasInternalType("foo"))).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[child::@InternalType='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestToken() {
	rule := On(HasToken("foo")).Roles(uast.SimpleIdentifier)
	expected := head + `| /self::\*\[@Token='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestAnd() {
	rule := On(And(
		HasToken("foo"),
		HasToken("bar"),
		HasInternalType("bla"),
	)).Roles(uast.SimpleIdentifier)
	expected := head +
		`| /self::\*\[\(@Token='foo'\) and \(@Token='bar'\) and \(@InternalType='bla'\)\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestOr() {
	rule := On(Or(
		HasToken("foo"),
		HasToken("bar"),
		HasInternalType("bla"),
	)).Roles(uast.SimpleIdentifier)
	expected := head +
		`| /self::\*\[\(@Token='foo'\) or \(@Token='bar'\) or \(@InternalType='bla'\)\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestSelf() {
	rule := On(Any).Self(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + `| /self::\*\[@Token='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)

	rule = On(Any).Self(On(HasToken("foo"))).Roles(uast.SimpleIdentifier)
	expected = head + `| /self::\*\[\*\] | SimpleIdentifier |` + "\n"
	obtained = rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestChildren() {
	rule := On(Any).Children(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + `| /\*\[@Token='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestDescendants() {
	rule := On(Any).Descendants(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + `| //\*\[@Token='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestDescendantsOrSelf() {
	rule := On(Any).DescendantsOrSelf(On(HasToken("foo")).Roles(uast.SimpleIdentifier))
	expected := head + `| /descendant\-or\-self::\*\[@Token='foo'\] | SimpleIdentifier |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestMisc1() {
	rule := On(Any).Self(
		On(Not(HasInternalType("FILE"))).Error(fmt.Errorf("root must be CompilationUnit")),
		On(HasInternalType("FILE")).Roles(uast.SimpleIdentifier).Descendants(
			On(HasInternalType("identifier")).Roles(uast.QualifiedIdentifier),
			On(HasInternalType("binary expression")).Children(
				On(HasInternalType("left")).Roles(uast.BinaryExpressionLeft)),
		))
	expected := head +
		`| /self::\*\[not\(@InternalType='FILE'\)\] | Error |` + "\n" +
		`| /self::\*\[@InternalType='FILE'\] | SimpleIdentifier |` + "\n" +
		`| /self::\*\[@InternalType='FILE'\]//\*\[@InternalType='identifier'\] | QualifiedIdentifier |` + "\n" +
		`| /self::\*\[@InternalType='FILE'\]//\*\[@InternalType='binary expression'\]/\*\[@InternalType='left'\] | BinaryExpressionLeft |` + "\n"
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}

func (suite *RulesDocSuite) TestMarkdownEscapes() {
	rule := On(Any).Descendants(
		On(HasInternalType(`\`)).Roles(uast.OpBooleanOr),
		On(HasInternalType("|")).Roles(uast.OpBooleanOr),
		On(HasInternalType("||")).Roles(uast.OpBooleanOr),
		On(HasInternalType("`")).Roles(uast.OpBooleanOr),
		On(HasInternalType("*")).Roles(uast.OpBooleanOr),
		On(HasInternalType("_")).Roles(uast.OpBooleanOr),
		On(HasInternalType("{")).Roles(uast.OpBooleanOr),
		On(HasInternalType("}")).Roles(uast.OpBooleanOr),
		On(HasInternalType("[")).Roles(uast.OpBooleanOr),
		On(HasInternalType("]")).Roles(uast.OpBooleanOr),
		On(HasInternalType("(")).Roles(uast.OpBooleanOr),
		On(HasInternalType(")")).Roles(uast.OpBooleanOr),
		On(HasInternalType("#")).Roles(uast.OpBooleanOr),
		On(HasInternalType("+")).Roles(uast.OpBooleanOr),
		On(HasInternalType("-")).Roles(uast.OpBooleanOr),
		On(HasInternalType(".")).Roles(uast.OpBooleanOr),
		On(HasInternalType("!")).Roles(uast.OpBooleanOr),
	)
	expected := head +
		`| //\*\[@InternalType='\\'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\|'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\|\|'\] | OpBooleanOr |` + "\n" +
		"| //\\*\\[@InternalType='`'\\] | OpBooleanOr |\n" +
		`| //\*\[@InternalType='\*'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\_'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\{'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\}'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\['\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\]'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\('\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\)'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\#'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\+'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\-'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\.'\] | OpBooleanOr |` + "\n" +
		`| //\*\[@InternalType='\!'\] | OpBooleanOr |` + "\n" +
		""
	obtained := rule.String()
	require.Equal(suite.T(), expected, obtained)
}
