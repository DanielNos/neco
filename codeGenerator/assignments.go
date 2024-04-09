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
			cg.addInstruction(VM.IT_StoreAndPop, cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier))
		} else {
			cg.addInstruction(VM.IT_Store, cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier))
		}

	// Assign to an object property
	case parser.NT_ObjectField:
		objectFieldNode := assignedTo.Value.(*parser.ObjectFieldNode)

		loadInstructionIndex := len(*cg.target)
		cg.generateExpression(objectFieldNode.Object)
		variableID := (*cg.target)[loadInstructionIndex].InstructionValue[0]

		cg.addInstruction(VM.IT_SetField, byte(objectFieldNode.FieldIndex))
		cg.addInstruction(VM.IT_StoreAndPop, variableID)

		if isLast {
			cg.addInstruction(VM.IT_Pop)
		}

	default:
		panic("Not implemented exception: CodeGenerator -> generateAssignmentInstruction for node" + assignedTo.NodeType.String())
	}
}
