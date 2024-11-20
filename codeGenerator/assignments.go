package codeGenerator

import (
	"github.com/DanielNos/neco/parser"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func (cg *CodeGenerator) storeTop(to *parser.Node, pop bool) {
	instruction := VM.IT_Store
	if pop {
		instruction = VM.IT_StoreAndPop
	}

	if to.NodeType == parser.NT_Variable {
		cg.addInstruction(instruction, cg.findVariableIdentifier(to.Value.(*parser.VariableNode).Identifier))
	} else if to.NodeType == parser.NT_ListValue {
		listAssignNode := to.Value.(*parser.TypedBinaryNode)
		cg.generateExpression(listAssignNode.Right)
		cg.addInstruction(VM.IT_SetListAtAToB, cg.findVariableIdentifier(listAssignNode.Left.Value.(*parser.VariableNode).Identifier))
	}
}

func (cg *CodeGenerator) generateAssignment(assignNode *parser.AssignNode) {
	// Check if all assigned to statements are variables
	if cg.optimize {
		noFieldAssigns := true
		for _, assignedTo := range assignNode.AssignedTo {
			if assignedTo.NodeType == parser.NT_ObjectField {
				noFieldAssigns = false
				break
			}
		}

		// If expression isn't assigned to any fields, reuse it
		if noFieldAssigns {
			cg.generateExpression(assignNode.AssignedExpression)

			for i, assignedTo := range assignNode.AssignedTo {
				if i == len(assignNode.AssignedTo)-1 {
					cg.storeTop(assignedTo, true)
				} else {
					cg.storeTop(assignedTo, false)
				}
			}
			return
		}
	}

	// Non-optimized assignments (expression regenerated for each assignment)
	for _, assignedTo := range assignNode.AssignedTo {

		// Assignment to a variable
		if assignedTo.NodeType == parser.NT_Variable {
			cg.generateExpression(assignNode.AssignedExpression)
			cg.addInstruction(VM.IT_StoreAndPop, cg.findVariableIdentifier(assignedTo.Value.(*parser.VariableNode).Identifier))

			// Assignment to an object field
		} else if assignedTo.NodeType == parser.NT_ObjectField {
			// Generate variable load and field getters
			startOfFields := len(*cg.target)
			cg.generateExpression(assignedTo)
			endOfFields := len(*cg.target)

			// Change field getter types from GetFieldAndPop to GetField
			for i := endOfFields - 1; i > startOfFields; i-- {
				(*cg.target)[i].InstructionType = VM.IT_GetField
			}

			cg.generateExpression(assignNode.AssignedExpression)

			// Copy field getters in reverse order and change them to field setters
			for i := endOfFields - 1; i > startOfFields; i-- {
				cg.addInstruction(VM.IT_SetField, (*cg.target)[i].InstructionValue[0])
			}

			(*cg.target)[endOfFields-1].InstructionType = IGNORE_INSTRUCTION

			cg.addInstruction(VM.IT_StoreAndPop, (*cg.target)[startOfFields].InstructionValue[0])

		} else {
			panic("Not implemented exception: CodeGenerator -> generateAssignmentInstruction for node" + assignedTo.NodeType.String())
		}
	}
}
