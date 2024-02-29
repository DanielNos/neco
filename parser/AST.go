package parser

import (
	"fmt"
	data "neco/dataStructures"
)

func Visualize(tree *Node) {
	moduleNode := tree.Value.(*ModuleNode)

	println(moduleNode.Name)

	for i, node := range moduleNode.Statements.Statements {
		visualize(node, "", i == len(moduleNode.Statements.Statements)-1)
	}
}

func VisualizeNode(node *Node) {
	visualize(node, "", true)
}

func visualize(node *Node, indent string, isLast bool) {
	fmt.Print(indent)

	if isLast {
		print("└─ ")
		indent += "   "
	} else {
		print("├─ ")
		indent += "│  "
	}

	switch node.NodeType {

	case NT_VariableDeclare:
		declare := node.Value.(*VariableDeclareNode)

		fmt.Printf("Declare %s:", declare.DataType)

		for _, id := range declare.Identifiers {
			fmt.Printf(" %s", id)
		}
		println()

	case NT_Assign:
		assign := node.Value.(*AssignNode)

		println("Assign")
		fmt.Printf("%s├─ %s\n", indent, assign.Identifier)

		visualize(assign.Expression, indent, true)

	case NT_Literal:
		literal := node.Value.(*LiteralNode)
		fmt.Printf("%s %v\n", literal.DType.String(), literal.Value)

	case NT_Add, NT_Subtract, NT_Multiply, NT_Divide, NT_Power, NT_Modulo,
		NT_Equal, NT_NotEqual, NT_Lower, NT_Greater, NT_LowerEqual, NT_GreaterEqual,
		NT_And, NT_Or:
		binary := node.Value.(*BinaryNode)
		fmt.Printf("%s (%s)\n", NodeTypeToString[node.NodeType], binary.DataType)

		if binary.Left != nil {
			visualize(binary.Left, indent, false)
		}

		visualize(binary.Right, indent, true)

	case NT_Not:
		println("!")
		visualize(node.Value.(*BinaryNode).Right, indent, true)

	case NT_Variable:
		println(node.Value.(*VariableNode).Identifier)

	case NT_FunctionDeclare:
		functionDeclareNode := node.Value.(*FunctionDeclareNode)

		fmt.Printf("fun %s(", functionDeclareNode.Identifier)

		if len(functionDeclareNode.Parameters) > 0 {
			fmt.Printf("%s %s", functionDeclareNode.Parameters[0].DataType, functionDeclareNode.Parameters[0].Identifier)
		}

		if len(functionDeclareNode.Parameters) > 1 {
			for _, parameter := range functionDeclareNode.Parameters[1:] {
				fmt.Printf(", %s %s", parameter.DataType, parameter.Identifier)
			}
		}

		print(") ")

		if functionDeclareNode.ReturnType.DType != data.DT_NoType {
			fmt.Printf("-> %s", functionDeclareNode.ReturnType)
		}

		fmt.Printf(" (%d)\n", functionDeclareNode.Number)

		scopeNode := functionDeclareNode.Body.Value.(*ScopeNode)

		for i, statement := range scopeNode.Statements {
			visualize(statement, indent, i == len(scopeNode.Statements)-1)
		}

	case NT_Scope:
		scopeNode := node.Value.(*ScopeNode)
		fmt.Printf("Scope %d\n", scopeNode.Id)

		for i, statement := range scopeNode.Statements {
			visualize(statement, indent, i == len(scopeNode.Statements)-1)
		}

	case NT_FunctionCall:
		functionCall := node.Value.(*FunctionCallNode)
		fmt.Printf("%s(...)\n", functionCall.Identifier)

		for i, argument := range functionCall.Arguments {
			visualize(argument, indent, i == len(functionCall.Arguments)-1)
		}

	case NT_Return:
		println("return")

		if node.Value != nil {
			visualize(node.Value.(*Node), indent, true)
		}

	case NT_If:
		println("if")
		ifNode := node.Value.(*IfNode)
		visualize(ifNode.IfStatements[0].Condition, indent, false)
		visualize(ifNode.IfStatements[0].Body, indent, len(ifNode.IfStatements) == 1 && ifNode.ElseBody == nil)

		if len(ifNode.IfStatements) == 1 {
			if ifNode.ElseBody != nil {
				visualize(ifNode.ElseBody, indent, true)
			}
		} else {
			for i, elif := range ifNode.IfStatements {
				visualize(elif.Condition, indent, i == len(ifNode.IfStatements)-1 && ifNode.ElseBody == nil)
				visualize(elif.Body, indent, i == len(ifNode.IfStatements)-1 && ifNode.ElseBody == nil)
			}
			if ifNode.ElseBody != nil {
				visualize(ifNode.ElseBody, indent, true)
			}
		}

	case NT_Loop:
		println("loop")
		visualize(node.Value.(*Node), indent, true)

	case NT_ForLoop:
		println("for")

		forNode := node.Value.(*ForLoopNode)

		fmt.Printf("%s├─ Init\n", indent)

		for i, node := range forNode.InitStatement {
			visualize(node, indent+"│  ", i == len(forNode.InitStatement)-1)
		}
		visualize(forNode.Body, indent, true)

	case NT_Break:
		println("break")

	case NT_List:
		listNode := node.Value.(*ListNode)
		fmt.Printf("%s\n", listNode.DataType)

		for i, node := range listNode.Nodes {
			visualize(node, indent, i == len(listNode.Nodes)-1)
		}

	case NT_ListValue:
		println("ListIndex")
		listValue := node.Value.(*BinaryNode)

		visualize(listValue.Left, indent, false)
		visualize(listValue.Right, indent, true)

	case NT_ListAssign:
		listAssign := node.Value.(*ListAssignNode)
		println("Assign")

		fmt.Printf("%s├─ %s[...]\n", indent, listAssign.Identifier)
		visualize(listAssign.IndexExpression, indent+"│  ", true)
		visualize(listAssign.AssignedExpression, indent, true)

	default:
		println(NodeTypeToString[node.NodeType])
	}
}
