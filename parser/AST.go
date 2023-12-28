package parser

import (
	"fmt"
)

func Visualize(tree *Node) {
	moduleNode := tree.value.(*ModuleNode)

	println(moduleNode.name)

	for i, node := range moduleNode.statements.statements {
		visualize(node, "", i == len(moduleNode.statements.statements)-1)
	}
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

	switch node.nodeType {

	case NT_VariableDeclare:
		declare := node.value.(*VariableDeclareNode)
		
		fmt.Printf("Declare %s", declare.dataType)

		if declare.canBeNone {
			print("?:")
		} else {
			print(":")
		}

		for _, id := range declare.identifiers {
			fmt.Printf(" %s", id)
		}
		println()

	case NT_Assign:
		assign := node.value.(*AssignNode)

		println("Assign")
		print(indent, "├─ ")

		for _, id := range assign.identifiers {
			fmt.Printf("%s ", id)
		}
		println()

		visualize(assign.expression, indent, true)

	case NT_Literal:
		literal := node.value.(*LiteralNode)
		println(literal.dataType.String(), literal.value)

	case NT_Add, NT_Subtract, NT_Multiply, NT_Divide, NT_Power, NT_Modulo,
		NT_Equal, NT_NotEqual, NT_Lower, NT_Greater, NT_LowerEqual, NT_GreaterEqual:
		println(NodeTypeToString[node.nodeType])
		binary := node.value.(*BinaryNode)

		visualize(binary.left, indent, false)
		visualize(binary.right, indent, true)

	case NT_Variable:
		println(node.value.(*VariableNode).identifier)

	case NT_FunctionDeclare:
		functionDeclareNode := node.value.(*FunctionDeclareNode)
	
		fmt.Printf("fun %s(", functionDeclareNode.identifier)

		if len(functionDeclareNode.parameters) > 0 {
			fmt.Printf("%s %s", functionDeclareNode.parameters[0].dataType, functionDeclareNode.parameters[0].identifier)
		}
		
		if len(functionDeclareNode.parameters) > 1 {
			for _, parameter := range functionDeclareNode.parameters[1:] {
				fmt.Printf(", %s %s", parameter.dataType, parameter.identifier)
			}
		}

		print(") ")

		if functionDeclareNode.returnType != nil {
			fmt.Printf("-> %s", functionDeclareNode.returnType)
		}

		println()

		scopeNode := functionDeclareNode.body.value.(*ScopeNode)

		for i, statement := range scopeNode.statements {
			visualize(statement, indent, i == len(scopeNode.statements) - 1)
		}

	case NT_Scope:
		println("Scope")
		scopeNode := node.value.(*ScopeNode)

		for i, statement := range scopeNode.statements {
			visualize(statement, indent, i == len(scopeNode.statements) - 1)
		}
	
	case NT_FunctionCall:
		functionCall := node.value.(*FunctionCallNode)
		fmt.Printf("%s(...)\n", functionCall.identifier)

		for i, argument := range functionCall.arguments {
			visualize(argument, indent, i == len(functionCall.arguments) - 1)
		}
		
	default:
		fmt.Printf("%s\n", NodeTypeToString[node.nodeType])
	}
}
