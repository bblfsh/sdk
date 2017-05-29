package uast

import (
	"errors"
)

//CodeReference
//https://pmd.github.io/pmd-5.7.0/pmd-java/xref/net/sourceforge/pmd/lang/java/rule/codesize/NPathComplexityRule.html

//NpathComplexity return a slice of int for each function in the tree
func NpathComplexity(n *Node) ([]int, error) {
	var funcs []*Node
	var npath []int

	if n.containsRol(FunctionDeclarationBody) {
		funcs = append(funcs, n)
	} else {
		funcs = n.deepChildrenOfRole(FunctionDeclarationBody)
	}
	if len(funcs) == 0 {
		npath = append(npath, -1)
		return npath, errors.New("Function declaration not found")
	}
	for _, function := range funcs {
		npath = append(npath, visitFunctionBody(function))
	}
	return npath, nil
}

func visitorSelector(n *Node) int {
	// I need to add a error when the node dont have any rol
	// when I got 2 or more roles that are inside the switch this doesn't work
	for _, rol := range n.Roles {
		switch rol {
		case If:
			return visitIf(n)
		case While:
			return visitWhile(n)
		case Switch:
			return visitSwitch(n)
		case DoWhile:
			return visitDoWhile(n)
		case For:
			return visitFor(n)
		case Return:
			return visitReturn(n)
		default:
		}
	}
	return visitNotCompNode(n)
}

func complexityMultOf(n *Node) int {
	npath := 1
	for _, child := range n.Children {
		npath *= visitorSelector(child)
	}
	return npath
}

func complexitySumOf(n *Node) int {
	npath := 0
	for _, child := range n.Children {
		npath += visitorSelector(child)
	}
	return npath
}

func visitFunctionBody(n *Node) int {
	return complexityMultOf(n)
}

func visitNotCompNode(n *Node) int {
	return complexityMultOf(n)
}

func visitIf(n *Node) int {
	// (npath of if + npath of else (or 1) + bool_comp of if) * npath of next
	npath := 0
	ifBody := n.childrenOfRole(IfBody)
	ifCondition := n.childrenOfRole(IfCondition)
	ifElse := n.childrenOfRole(IfElse)

	if len(ifElse) == 0 {
		npath++
	} else {
		//This if is a short circuit to avoid the two roles in the switch problem
		if ifElse[0].containsRol(If) {
			npath += visitIf(ifElse[0])
		} else {
			npath += complexityMultOf(ifElse[0])
		}
	}
	npath *= complexityMultOf(ifBody[0])
	npath += expressionComp(ifCondition[0])

	return npath
}

func visitWhile(n *Node) int {
	// (npath of while + bool_comp of while + npath of else (or 1)) * npath of next
	npath := 0
	whileCondition := n.childrenOfRole(WhileCondition)
	whileBody := n.childrenOfRole(WhileBody)
	whileElse := n.childrenOfRole(IfElse)
	//Some languages like python can have an else in a while loop
	if len(whileElse) == 0 {
		npath++
	} else {
		npath += complexityMultOf(whileElse[0])
	}
	npath *= complexityMultOf(whileBody[0])
	npath += expressionComp(whileCondition[0])

	return npath
}

func visitDoWhile(n *Node) int {
	// (npath of do + bool_comp of do + 1) * npath of next
	npath := 1
	doWhileCondition := n.childrenOfRole(DoWhileCondition)
	doWhileBody := n.childrenOfRole(DoWhileBody)

	npath *= complexityMultOf(doWhileBody[0])
	npath += expressionComp(doWhileCondition[0])

	//The +1 is used for the path of not taking the doWhile
	return npath + 1
}

func visitFor(n *Node) int {
	// (npath of for + bool_comp of for + 1) * npath of next
	npath := 1
	forBody := n.childrenOfRole(ForBody)
	//forExpression := n.childrenOfRole(ForExpression)
	npath *= complexityMultOf(forBody[0])
	//This is suposed the way of doing, but I cant find a example that works with pmd, for pmd the value is 1
	//npath += expressionComp(forExpression[0])
	npath++
	return npath
}

func visitReturn(n *Node) int {
	if aux := expressionComp(n); aux != 1 {
		return aux - 1
	}
	return 1
}

func visitSwitch(n *Node) int {

	caseDefault := n.childrenOfRole(SwitchDefault)
	switchCases := n.childrenOfRole(SwitchCase)
	switchCondition := n.childrenOfRole(SwitchCaseCondition)
	npath := 0
	//In pmd the expressionComp function returns always our value -1
	//but in other places of the code the fuction works exactly as our function
	//I suposed this happens because java AST differs with the UAST
	npath += expressionComp(switchCondition[0]) - 1
	if len(caseDefault) != 0 {
		npath += complexityMultOf(caseDefault[0])
	}
	for _, switchCase := range switchCases {
		npath += complexityMultOf(switchCase)
	}
	return npath
}

func visitTry(n *Node) {
	//TODO, in the code of reference it isn't impelemted yet
}

func visitConditionalExpr(n *Node) {
	//TODO ternary operators are not defined on the UAST yet
}

func (n *Node) childrenOfRole(wanted Role) []*Node {
	var children []*Node
	for _, child := range n.Children {
		for _, rol := range child.Roles {
			if rol == wanted {
				children = append(children, child)
			}
		}
	}
	return children
}

func (n *Node) deepChildrenOfRole(rol Role) []*Node {
	var childList []*Node
	for _, child := range n.Children {
		for _, cRol := range child.Roles {
			if cRol == rol {
				childList = append(childList, child)
			}
			childList = append(childList, child.deepChildrenOfRole(rol)...)
		}
	}
	return childList
}

func expressionComp(n *Node) int {
	orCount := n.deepCountChildrenOfRol(OpBooleanAnd)
	andCount := n.deepCountChildrenOfRol(OpBooleanOr)

	return orCount + andCount + 1
}

func (n *Node) deepCountChildrenOfRol(rol Role) int {
	count := 0
	for _, child := range n.Children {
		for _, cRol := range child.Roles {
			if cRol == rol {
				count++
			}
			count += child.deepCountChildrenOfRol(rol)
		}
	}
	return count
}

func (n *Node) containsRol(rol Role) bool {
	for _, r := range n.Roles {
		if rol == r {
			return true
		}
	}
	return false
}
