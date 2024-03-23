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
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StoreAndPop, []byte{cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier)}})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Store, []byte{cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier)}})
		}
	}
}