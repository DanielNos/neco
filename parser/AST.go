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
		visualize(assign.AssignedExpression, indent, false)
		if len(assign.AssignedTo) > 1 {
			println(indent + "└─ [multiple]")
			indent += "   "

			for i, node := range assign.AssignedTo {
				visualize(node, indent, i == len(assign.AssignedTo)-1)
			}
		} else {
			visualize(assign.AssignedTo[len(assign.AssignedTo)-1], indent, true)
		}

	case NT_Literal:
		literal := node.Value.(*LiteralNode)
		if literal.PrimitiveType == data.DT_String {
			fmt.Printf("%s \"%s\"\n", literal.PrimitiveType.String(), literal.Value)
		} else {
			fmt.Printf("%s %v\n", literal.PrimitiveType.String(), literal.Value)
		}

	case NT_Add, NT_Subtract, NT_Multiply, NT_Divide, NT_Power, NT_Modulo,
		NT_Equal, NT_NotEqual, NT_Lower, NT_Greater, NT_LowerEqual, NT_GreaterEqual,
		NT_And, NT_Or, NT_In:
		binary := node.Value.(*TypedBinaryNode)
		fmt.Printf("%s (%s)\n", NodeTypeToString[node.NodeType], binary.DataType)

		if binary.Left != nil {
			visualize(binary.Left, indent, false)
		}

		visualize(binary.Right, indent, true)

	case NT_Not:
		println("!")
		visualize(node.Value.(*TypedBinaryNode).Right, indent, true)

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

		if functionDeclareNode.ReturnType.Type != data.DT_NoType {
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

	case NT_ForEachLoop:
		println("forEach")
		forEach := node.Value.(*ForEachLoopNode)

		visualize(forEach.Iterator, indent, false)
		visualize(forEach.IteratedExpression, indent, false)
		visualize(forEach.Body, indent, true)

	case NT_Break:
		println("break")

	case NT_List, NT_Set:
		listNode := node.Value.(*ListNode)
		fmt.Printf("%s", listNode.DataType)

		if len(listNode.Nodes) == 0 {
			println(" (empty)")
			return
		} else {
			println()
		}

		for i, node := range listNode.Nodes {
			visualize(node, indent, i == len(listNode.Nodes)-1)
		}

	case NT_ListValue:
		println("ListIndex")
		listValue := node.Value.(*TypedBinaryNode)

		visualize(listValue.Left, indent, false)
		visualize(listValue.Right, indent, true)

	case NT_ListAssign:
		listAssign := node.Value.(*ListAssignNode)
		println("Assign")

		fmt.Printf("%s├─ %s[...]\n", indent, listAssign.Identifier)
		visualize(listAssign.IndexExpression, indent+"│  ", true)
		visualize(listAssign.AssignedExpression, indent, true)

	case NT_Enum:
		fmt.Printf("%d (%s)\n", node.Value.(*EnumNode).Value, node.Value.(*EnumNode).Identifier)

	case NT_Object:
		ObjectNode := node.Value.(*ObjectNode)

		fmt.Printf("%s\n", ObjectNode.Identifier)

		for i, n := range ObjectNode.Properties {
			visualize(n, indent, i == len(ObjectNode.Properties)-1)
		}

	case NT_Delete:
		println("Delete")
		visualize(node.Value.(*Node), indent, true)

	default:
		println(NodeTypeToString[node.NodeType])
	}
}
