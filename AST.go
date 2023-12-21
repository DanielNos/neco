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
		
		fmt.Printf("Declare %s:", declare.dataType)

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
	
	default:
		println("???")
	}
}
