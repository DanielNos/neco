package codeGenerator

import (
	data "neco/dataStructures"
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateVariableDeclaration(node *parser.Node) {
	variable := node.Value.(*parser.VariableDeclareNode)

	for i := 0; i < len(variable.Identifiers); i++ {
		id, isRedeclared := cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]]

		// Variable with this identifier is declared for the first time
		if !isRedeclared {
			cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]] = cg.variableIdentifierCounters.Top.Value.(uint8)

			id = cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]]

			cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1
		}

		cg.generateVariableDeclarator(variable.DataType, &id)
	}
}

func (cg *CodeGenerator) generateVariableDeclarator(dataType *data.DataType, id *uint8) {

	// Generate declaration of root type
	if id != nil {
		cg.addInstruction(dataTypeToDeclareInstruction[dataType.Type], *id)
	} else {
		// Identifier of variable (id) is passed only for root types, sub-types have no arguments
		cg.addInstruction(dataTypeToDeclareInstruction[dataType.Type])

	}

	// Generate sub-type of composite types
	if dataType.SubType != nil && (dataType.Type == data.DT_List || dataType.Type == data.DT_Set) {
		cg.generateVariableDeclarator(dataType.SubType.(*data.DataType), nil)
	}
}

func (cg *CodeGenerator) generateDeletion(target *parser.Node) {
	switch target.NodeType {
	case parser.NT_Variable:
		// We don't actually delete anything, variable is redeclared with the same identifier

	case parser.NT_In:
		inNode := target.Value.(*parser.TypedBinaryNode)

		// Only generate element removal if set isn't a literal
		if inNode.Right.NodeType == parser.NT_Variable {
			cg.generateExpression(inNode.Right) // Generate set
			cg.generateExpression(inNode.Left)  // Generate element

			// Remove it
			cg.addInstruction(VM.IT_RemoveSetElement)

			// Generate set load
			cg.generateExpression(inNode.Right)
			// Replace set load instruction type with StoreAndPop
			(*cg.target)[len(*cg.target)-1].InstructionType = VM.IT_StoreAndPop
		}
	}
}
