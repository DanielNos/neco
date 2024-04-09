package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateAssignment(assignNode *parser.AssignNode) {
	cg.generateExpression(assignNode.AssignedExpression)

	for i, assignedTo := range assignNode.AssignedTo {
		cg.generateAssignmentInstruction(assignedTo, i == len(assignNode.AssignedTo)-1 && cg.optimize)
	}
}

func (cg *CodeGenerator) generateAssignmentInstruction(assignedTo *parser.Node, isLast bool) {
	switch assignedTo.NodeType {
	// Assign to a variable
	case parser.NT_Variable:
		if isLast {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_StoreAndPop, []byte{cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier)}})
		} else {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_Store, []byte{cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier)}})
		}

	// Assign to an object property
	case parser.NT_ObjectField:
		objectFieldNode := assignedTo.Value.(*parser.ObjectFieldNode)

		loadInstructionIndex := len(*cg.target)
		cg.generateExpression(objectFieldNode.Object)
		variableID := (*cg.target)[loadInstructionIndex].InstructionValue[0]

		*cg.target = append(*cg.target, VM.Instruction{VM.IT_SetField, []byte{byte(objectFieldNode.FieldIndex)}})
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_StoreAndPop, []byte{variableID}})

		if isLast {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_Pop, NO_ARGS})
		}

	default:
		panic("Not implemented exception: CodeGenerator -> generateAssignmentInstruction for node" + assignedTo.NodeType.String())
	}
}
