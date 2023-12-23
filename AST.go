package main

import "fmt"

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
		
	default:
		println("???")
	}
}
