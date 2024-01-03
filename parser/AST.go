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
		
		fmt.Printf("Declare %s", declare.variableType)

		if declare.variableType.canBeNone {
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
		fmt.Printf("%s├─ %s\n", indent, assign.identifier)

		visualize(assign.expression, indent, true)

	case NT_Literal:
		literal := node.value.(*LiteralNode)
		fmt.Printf("%s %v\n", literal.dataType.String(), literal.value)

	case NT_Add, NT_Subtract, NT_Multiply, NT_Divide, NT_Power, NT_Modulo,
		NT_Equal, NT_NotEqual, NT_Lower, NT_Greater, NT_LowerEqual, NT_GreaterEqual,
		NT_And, NT_Or:
		println(NodeTypeToString[node.nodeType])
		binary := node.value.(*BinaryNode)

		if binary.left != nil {
			visualize(binary.left, indent, false)
		}

		visualize(binary.right, indent, true)
	
	case NT_Not:
		println("!")
		visualize(node.value.(*BinaryNode).right, indent, true)

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

		if functionDeclareNode.returnType.dataType != DT_NoType {
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

	case NT_Return:
		println("return")

		if node.value != nil {
			visualize(node.value.(*Node), indent, true)
		}

	case NT_If:
		println("if")
		ifNode := node.value.(*IfNode)
		visualize(ifNode.condition, indent, false)
		visualize(ifNode.body, indent, len(ifNode.elseIfs) == 0 && ifNode.elseBody == nil)

		if len(ifNode.elseIfs) == 0 {
			if ifNode.elseBody != nil {
				visualize(ifNode.elseBody, indent, true)
			}
		} else {
			for i, elif := range ifNode.elseIfs {
				visualize(elif, indent, i == len(ifNode.elseIfs) - 1 && ifNode.elseBody == nil)
			}
			if ifNode.elseBody != nil {
				visualize(ifNode.elseBody, indent, true)
			}
		}

	default:
		fmt.Printf("%s\n", NodeTypeToString[node.nodeType])
	}
}
