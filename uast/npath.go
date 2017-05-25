package uast

import (
	"errors"
	"fmt"
)

//CodeReference
//https://pmd.github.io/pmd-5.7.0/pmd-java/xref/net/sourceforge/pmd/lang/java/rule/codesize/NPathComplexityRule.html

//I don't know what is better a map Rol-> func or a switch
//var selector = map[Role]func(n *Node) int{}

func NpathComplexity(n *Node) ([]int, error) {
	var funcs []*Node
	var npath []int
	if n.containsRol(FunctionDeclarationBody) {
		funcs = append(funcs, n)
	} else {
		funcs = n.deepChildrenOfRole(FunctionDeclarationBody)
	}
	if len(funcs) <= 0 {
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
	// when I got 2 or more roles in a node this doesnt work
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
			return visitNotCompNode(n)
		}
	}
	return -10
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
		npath += complexityMultOf(ifElse[0])
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
		fmt.Println("WhileELse ", npath)
	}
	npath *= complexityMultOf(whileBody[0])
	npath += expressionComp(whileCondition[0])

	return npath
}

func visitDoWhile(n *Node) int {
	// (npath of do + bool_comp of do + 1) * npath of next
	npath := 0
	doWhileCondition := n.childrenOfRole(DoWhileCondition)
	doWhileBody := n.childrenOfRole(DoWhileBody)

	npath += complexitySumOf(doWhileBody[0])
	npath += expressionComp(doWhileCondition[0])
	//The +1 is used for the path of not taking the doWhile
	return npath + 1
}

func visitFor(n *Node) int {
	// (npath of for + bool_comp of for + 1) * npath of next
	npath := 0
	forBody := n.childrenOfRole(ForBody)
	forExpression := n.childrenOfRole(ForExpression)

	npath += complexitySumOf(forBody[0])
	npath += expressionComp(forExpression[0])

	return npath + 1
}

func visitReturn(n *Node) int {
	//The return isn't complete, I don't fully understand what PMD people do here
	return expressionComp(n)
}

func visitSwitch(n *Node) int {
	// The switch npath calculation is strange too in PMD
	npath := 0
	switchCases := n.childrenOfRole(SwitchCase)
	switchDefault := n.childrenOfRole(SwitchDefault)

	for _, switchCase := range switchCases {
		npath = complexityMultOf(switchCase)
	}
	if len(switchDefault) == 0 {
		npath++
	} else {
		npath += complexityMultOf(switchDefault[0])
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
