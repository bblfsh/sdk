// Code generated by "stringer -type=Role"; DO NOT EDIT

package uast

import "fmt"

const _Role_name = "SimpleIdentifierQualifiedIdentifierExpressionStatementFilePackageDeclarationImportDeclarationImportPathImportAliasFunctionDeclarationTypeDeclarationStaticVisibleFromInstanceVisibleFromTypeVisibleFromSubtypeVisibleFromPackageVisibleFromSubpackageVisibleFromModuleVisibleFromFriendVisibleFromWorldIfIfConditionIfBodyIfElseSwitchSwitchCaseSwitchCaseConditionSwitchCaseBodySwitchDefaultForForInitForExpressionForUpdateForBodyForEachWhileWhileConditionWhileBodyDoWhileDoWhileConditionDoWhileBodyBreakContinueBlockBlockScopeReturnTryTryBodyTryCatchTryFinallyThrowAssertMethodInvocationMethodInvocationObjectMethodInvocationNameMethodInvocationArgumentNoopLiteralNullLiteralStringLiteralNumberLiteralTypeLiteralTypePrimitiveTypeAssignmentAssignmentVariableAssignmentValueThisCommentDocumentationWhitespace"

var _Role_index = [...]uint16{0, 16, 35, 45, 54, 58, 76, 93, 103, 114, 133, 148, 154, 173, 188, 206, 224, 245, 262, 279, 295, 297, 308, 314, 320, 326, 336, 355, 369, 382, 385, 392, 405, 414, 421, 428, 433, 447, 456, 463, 479, 490, 495, 503, 508, 518, 524, 527, 534, 542, 552, 557, 563, 579, 601, 621, 645, 649, 656, 667, 680, 693, 704, 708, 721, 731, 749, 764, 768, 775, 788, 798}

func (i Role) String() string {
	if i < 0 || i >= Role(len(_Role_index)-1) {
		return fmt.Sprintf("Role(%d)", i)
	}
	return _Role_name[_Role_index[i]:_Role_index[i+1]]
}
