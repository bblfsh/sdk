package uast

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	result := n1.deepCountChildrenOfRol(Statement)
	expect := 2
	require.Equal(expect, result)

	result = n2.deepCountChildrenOfRol(Statement)
	expect = 3
	require.Equal(expect, result)
}

func TestFindChildrenOfRol(t *testing.T) {
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

	result := n1.childrenOfRole(Statement)
	expect := 2
	require.Equal(expect, len(result))

	result = n2.childrenOfRole(Statement)
	expect = 1
	require.Equal(expect, len(result))

	result = n1.deepChildrenOfRole(Statement)
	expect = 2
	require.Equal(expect, len(result))

	result = n2.deepChildrenOfRole(Statement)
	expect = 3
	require.Equal(expect, len(result))
}

func TestExpresionComplex(t *testing.T) {
	require := require.New(t)

	n := &Node{InternalType: "ifCondition", Roles: []Role{IfCondition}, Children: []*Node{
		{InternalType: "bool_and", Roles: []Role{OpBooleanAnd}},
		{InternalType: "bool_xor", Roles: []Role{OpBooleanXor}},
	}}
	n2 := &Node{InternalType: "ifCondition", Roles: []Role{IfCondition}, Children: []*Node{
		{InternalType: "bool_and", Roles: []Role{OpBooleanAnd}, Children: []*Node{
			{InternalType: "bool_or", Roles: []Role{OpBooleanOr}, Children: []*Node{
				{InternalType: "bool_xor", Roles: []Role{OpBooleanXor}},
			}},
		}},
	}}

	result := expressionComp(n)
	expect := 2
	require.Equal(expect, result)

	result = expressionComp(n2)
	expect = 3
	require.Equal(expect, result)
}

func TestNpathComplexity(t *testing.T) {
	require := require.New(t)
	var result []int
	var expect []int

	andBool := &Node{InternalType: "bool_and", Roles: []Role{OpBooleanAnd}}
	orBool := &Node{InternalType: "bool_or", Roles: []Role{OpBooleanOr}}
	statement := &Node{InternalType: "Statement", Roles: []Role{Statement}}

	n := &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		statement,
	}}

	comp, err := NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 1)
	/*
			if(3conditions){
				Statement
				Statement
			}else if(3conditions){
				Statement
				Statement
		  }else{
				Statement
				Statement
		  } Npath = 7
	*/
	ifCondition := &Node{InternalType: "Condition", Roles: []Role{IfCondition}, Children: []*Node{
		andBool,
		orBool,
	}}
	ifBody := &Node{InternalType: "Body", Roles: []Role{IfBody}, Children: []*Node{
		statement,
		statement,
	}}
	elseIf := &Node{InternalType: "elseIf", Roles: []Role{IfElse}, Children: []*Node{
		&Node{InternalType: "If", Roles: []Role{If}, Children: []*Node{
			ifCondition,
			ifBody,
		}},
	}}
	ifElse := &Node{InternalType: "else", Roles: []Role{IfElse}, Children: []*Node{
		ifBody,
	}}
	nIf := &Node{InternalType: "if", Roles: []Role{If}, Children: []*Node{
		ifCondition,
		ifBody,
		elseIf,
		ifElse,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nIf,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 7)

	//This case looks like the previous one, but we have the ElseIF and the If roles in the same node
	/*
		if(3conditions){
			Statement
			Statement
		}else if(3conditions){
			Statement
			Statement
		}else{
			Statement
			Statement
		} Npath = 7
	*/

	elseIf2roles := &Node{InternalType: "elseIf", Roles: []Role{IfElse, If}, Children: []*Node{
		ifCondition,
		ifBody,
	}}

	nIf2roles := &Node{InternalType: "if", Roles: []Role{If}, Children: []*Node{
		ifCondition,
		ifBody,
		elseIf2roles,
		ifElse,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nIf2roles,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 7)

	/*
	  if(condition){
	    Statement
	    Statement
	  }Npath = 2
	*/
	nSimpleIF := &Node{InternalType: "If", Roles: []Role{If}, Children: []*Node{
		{InternalType: "ifCondition", Roles: []Role{IfCondition}, Children: []*Node{}},
		ifBody,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nSimpleIF,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 2)

	/*
		The same if structure of the example above
		but repeated three times, in sequencial way
		Npath = 343
	*/

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nIf,
		nIf,
		nIf,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 343)

	/*
		if(3conditions){
			if(3conditions){
				if(3conditions){
					Statement
					Statemenet
				}else{
					Statement
					Statement
				}
			}
		} Npath = 10
	*/
	nestedIfBody := &Node{InternalType: "bodyº", Roles: []Role{IfBody}, Children: []*Node{
		{InternalType: "if2", Roles: []Role{If}, Children: []*Node{
			ifCondition,
			{InternalType: "body2", Roles: []Role{IfBody}, Children: []*Node{
				{InternalType: "if3", Roles: []Role{If}, Children: []*Node{
					ifCondition,
					ifBody,
					ifElse,
				}},
			}},
		}},
	}}
	nNestedIf := &Node{InternalType: "if1", Roles: []Role{If}, Children: []*Node{
		ifCondition,
		nestedIfBody,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nNestedIf,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 10)

	/*
		while(2condtions){
			Statement
			Statement
			Statement
		}else{
			Statement
			Statement
		} Npath = 3
	*/
	whileCondition := &Node{InternalType: "WhileCondition", Roles: []Role{WhileCondition}, Children: []*Node{
		andBool,
	}}
	whileBody := &Node{InternalType: "WhileBody", Roles: []Role{WhileBody}, Children: []*Node{
		statement,
		statement,
		statement,
	}}
	whileElse := &Node{InternalType: "WhileElse", Roles: []Role{IfElse}, Children: []*Node{
		statement,
		statement,
	}}
	nWhile := &Node{InternalType: "While", Roles: []Role{While}, Children: []*Node{
		whileCondition,
		whileBody,
		whileElse,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nWhile,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 3)

	/*
		while(2conditions){
			while(2conditions){
				while(2conditions){
					Statement
					Statement
				}
			}
		} Npath = 7
	*/
	nestedWhileBody := &Node{InternalType: "WhileBody1", Roles: []Role{WhileBody}, Children: []*Node{
		{InternalType: "While2", Roles: []Role{While}, Children: []*Node{
			whileCondition,
			{InternalType: "WhileBody2", Roles: []Role{WhileBody}, Children: []*Node{
				{InternalType: "While3", Roles: []Role{While}, Children: []*Node{
					whileCondition,
					whileBody,
				}},
			}},
		}},
	}}
	nNestedWhile := &Node{InternalType: "While1", Roles: []Role{While}, Children: []*Node{
		whileCondition,
		nestedWhileBody,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nNestedWhile,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 7)

	/*
			 for(init;2condition;update){
			 	Statement
				Statement
		 	 } Npath = 2
	*/
	forCondition := &Node{InternalType: "forCondition", Roles: []Role{ForExpression}, Children: []*Node{
		orBool,
	}}
	forInit := &Node{InternalType: "forInit", Roles: []Role{ForInit}}
	forUpdate := &Node{InternalType: "forUpdate", Roles: []Role{ForUpdate}}
	forBody := &Node{InternalType: "forBody", Roles: []Role{ForBody}, Children: []*Node{
		statement,
		statement,
	}}
	nFor := &Node{InternalType: "for", Roles: []Role{For}, Children: []*Node{
		forInit,
		forCondition,
		forUpdate,
		forBody,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nFor,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 2)

	/*
		for(init;2conditions;update){
			for(init;2conditions;update){
				for(init;2condtions;update){
					Statement
					Statement
				}
			}
		} Npath = 4
	*/
	nestedForBody := &Node{InternalType: "forBody1", Roles: []Role{ForBody}, Children: []*Node{
		{InternalType: "for2", Roles: []Role{For}, Children: []*Node{
			forInit,
			forCondition,
			forUpdate,
			{InternalType: "forBody2", Roles: []Role{ForBody}, Children: []*Node{
				{InternalType: "for3", Roles: []Role{For}, Children: []*Node{
					forInit,
					forCondition,
					forUpdate,
					forBody,
				}},
			}},
		}},
	}}
	nNestedFor := &Node{InternalType: "for1", Roles: []Role{For}, Children: []*Node{
		forInit,
		forCondition,
		forUpdate,
		nestedForBody,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nNestedFor,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 4)

	/*
		do{
			Statement
			Statement
		}while(3conditions)
		Npath = 4
	*/
	doWhileCondition := &Node{InternalType: "doWhileCondition", Roles: []Role{DoWhileCondition}, Children: []*Node{
		orBool,
	}}
	doWhileBody := &Node{InternalType: "doWhileBody", Roles: []Role{DoWhileBody}, Children: []*Node{
		statement,
		statement,
	}}
	nDoWhile := &Node{InternalType: "doWhile", Roles: []Role{DoWhile}, Children: []*Node{
		doWhileBody,
		doWhileCondition,
	}}
	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nDoWhile,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 4)

	/*
		do{
			do{
				do{
					Statement
					Statement
				}while(3conditions)
			}while{3conditions}
		}while(3condtions)
		Npath = 10
	*/
	nestedDoWhileBody := &Node{InternalType: "doWhileBody1", Roles: []Role{DoWhileBody}, Children: []*Node{
		{InternalType: "doWhile2", Roles: []Role{DoWhile}, Children: []*Node{
			{InternalType: "doWhileBody2", Roles: []Role{DoWhileBody}, Children: []*Node{
				{InternalType: "doWhile3", Roles: []Role{DoWhile}, Children: []*Node{
					doWhileBody,
					doWhileCondition,
				}},
			}},
			doWhileCondition,
		}},
	}}
	nNestedDoWhile := &Node{InternalType: "doWhile1", Roles: []Role{DoWhile}, Children: []*Node{
		nestedDoWhileBody,
		doWhileCondition,
	}}
	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nNestedDoWhile,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 10)

	/*
		switch(3conditions){
		case:
			Statement
			Statement
		case:
			Statement
			Statement
		default:
			Statement
			Statement
		} Npath = 5
	*/
	switchCondition := &Node{InternalType: "switchCondition", Roles: []Role{SwitchCaseCondition}, Children: []*Node{
		orBool,
		andBool,
	}}
	switchCaseBody := &Node{InternalType: "switchCaseBody", Roles: []Role{SwitchCaseBody}, Children: []*Node{
		statement,
		statement,
	}}
	switchCase := &Node{InternalType: "switchCase", Roles: []Role{SwitchCase}, Children: []*Node{
		switchCaseBody,
	}}
	defaultCase := &Node{InternalType: "defaultCase", Roles: []Role{SwitchDefault}, Children: []*Node{
		switchCaseBody,
	}}
	nSwitch := &Node{InternalType: "switch", Roles: []Role{Switch}, Children: []*Node{
		switchCondition,
		switchCase,
		switchCase,
		defaultCase,
	}}
	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nSwitch,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 5)

	/*
		switch(3conditions){
		case:
			Statement
			Statement
		case:
			Statement
			Statement
		default:
			switch(3conditions){
			case:
				Statement
				Statement
			case:
				Statement
				Statement
			default:
				Statement
				Statement
		} Npath = 9
	*/
	nestedDefaultCase := &Node{InternalType: "defaultCase", Roles: []Role{SwitchDefault}, Children: []*Node{
		nSwitch,
	}}
	nNestedSwitch := &Node{InternalType: "switch", Roles: []Role{Switch}, Children: []*Node{
		switchCondition,
		switchCase,
		switchCase,
		nestedDefaultCase,
	}}

	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		nNestedSwitch,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 9)

	/*
		return
	*/
	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		{InternalType: "Return", Roles: []Role{Return}},
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 1)

	/*
		statement
		statement
		return 3condition
	*/
	nReturn := &Node{InternalType: "Return", Roles: []Role{Return}, Children: []*Node{
		orBool,
		andBool,
	}}
	n = &Node{InternalType: "Function declaration body", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{
		statement,
		statement,
		nReturn,
	}}

	comp, err = NpathComplexity(n)
	result = append(result, comp...)
	if err != nil {
		fmt.Println(err)
	}
	expect = append(expect, 2)

	require.Equal(expect, result)
}

func TestNpathMultiFunc(t *testing.T) {
	require := require.New(t)
	var result []int
	expect := []int{7, 7, 7}
	andBool := &Node{InternalType: "bool_and", Roles: []Role{OpBooleanAnd}}
	orBool := &Node{InternalType: "bool_or", Roles: []Role{OpBooleanOr}}
	statement := &Node{InternalType: "Statement", Roles: []Role{Statement}}

	ifCondition := &Node{InternalType: "Condition", Roles: []Role{IfCondition}, Children: []*Node{
		andBool,
		orBool,
	}}
	ifBody := &Node{InternalType: "Body", Roles: []Role{IfBody}, Children: []*Node{
		statement,
		statement,
	}}
	elseIf := &Node{InternalType: "elseIf", Roles: []Role{IfElse}, Children: []*Node{
		&Node{InternalType: "If", Roles: []Role{If}, Children: []*Node{
			ifCondition,
			ifBody,
		}},
	}}
	ifElse := &Node{InternalType: "else", Roles: []Role{IfElse}, Children: []*Node{
		ifBody,
	}}
	nIf := &Node{InternalType: "if", Roles: []Role{If}, Children: []*Node{
		ifCondition,
		ifBody,
		elseIf,
		ifElse,
	}}

	func1 := &Node{InternalType: "func1", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{nIf}}
	func2 := &Node{InternalType: "func2", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{nIf}}
	func3 := &Node{InternalType: "func3", Roles: []Role{FunctionDeclarationBody}, Children: []*Node{nIf}}

	n := &Node{InternalType: "module", Children: []*Node{
		func1,
		func2,
		func3,
	}}
	comp, error := NpathComplexity(n)
	if error != nil {
		fmt.Println(error)
	}

	result = comp
	require.Equal(expect, result)
}
func TestZeroFunction(t *testing.T) {
	//Empty tree
	n := &Node{InternalType: "module"}
	_, error := NpathComplexity(n)
	assert.NotNil(t, error)
}
