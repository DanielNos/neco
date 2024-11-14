package codeGenerator

import (
	data "github.com/DanielNos/neco/dataStructures"
	"github.com/DanielNos/neco/parser"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func (cg *CodeGenerator) generateVariableDeclaration(node *parser.Node) {
	variable := node.Value.(*parser.VariableDeclareNode)

	for i := 0; i < len(variable.Identifiers); i++ {
		id, isRedeclared := cg.scopes.Top.Value.(*Scope).variableIdentifiers[variable.Identifiers[i]]

		// Variable with this identifier is declared for the first time
		if !isRedeclared {
			cg.scopes.Top.Value.(*Scope).variableIdentifiers[variable.Identifiers[i]] = cg.scopes.Top.Value.(*Scope).variableIdentifierCounter

			id = cg.scopes.Top.Value.(*Scope).variableIdentifiers[variable.Identifiers[i]]

			cg.scopes.Top.Value.(*Scope).variableIdentifierCounter++
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

	case parser.NT_ListValue:
		inNode := target.Value.(*parser.TypedBinaryNode)

		// Only generate element removal if set isn't a literal
		if inNode.Left.NodeType == parser.NT_Variable {
			cg.generateExpression(inNode.Left)  // Generate set
			cg.generateExpression(inNode.Right) // Generate element

			// Remove it
			if inNode.Left.Value.(*parser.VariableNode).DataType.Type == data.DT_Set {
				cg.addInstruction(VM.IT_RemoveSetElement)
			} else {
				cg.addInstruction(VM.IT_RemoveListElement)
			}

			// Generate set load
			cg.generateExpression(inNode.Left)

			// Replace set load instruction type with StoreAndPop
			(*cg.target)[len(*cg.target)-1].InstructionType = VM.IT_StoreAndPop
		}
	}
}
