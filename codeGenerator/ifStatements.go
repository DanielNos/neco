package codeGenerator

import (
	"github.com/DanielNos/neco/parser"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func (cg *CodeGenerator) generateIfStatement(ifStatement *parser.IfNode) {
	// Create slice for jump instructions and their positions so their destination can be set later
	jumpInstructions := make([]*VM.Instruction, len(ifStatement.IfStatements))
	jumpInstructionPositions := make([]int, len(ifStatement.IfStatements))

	// Generate else if conditions and jumps
	for i, statement := range ifStatement.IfStatements {
		// Generate condition expression
		cg.generateExpression(statement.Condition)
		// Generate conditional jump
		cg.addInstruction(VM.IT_JumpIfTrue, 0)

		// Store the jump instruction and it's position so destination position can be set later
		jumpInstructions[i] = &(*cg.target)[len(*cg.target)-1]
		jumpInstructionPositions[i] = len(*cg.target)
	}

	// Generate else body
	if ifStatement.ElseBody != nil {
		cg.generateScope(ifStatement.ElseBody.Value.(*parser.ScopeNode), nil)
	}

	// Generate jump instruction that will jump over all elifs
	cg.addInstruction(VM.IT_Jump, 0)

	// Store the jump instruction so it's destination can  be set later
	jumpFromElse := &(*cg.target)[len(*cg.target)-1]

	// Store jump position
	jumpFromElsePosition := len(*cg.target)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set if's conditional jump destination to next instruction
		updateJumpDistance(jumpInstructions[i], len(*cg.target)-jumpInstructionPositions[i], VM.IT_JumpIfTrueEx)

		// Generate if's body
		cg.generateScope(statement.Body.Value.(*parser.ScopeNode), nil)

		// Generate jump instruction to the end of if bodies
		cg.addInstruction(VM.IT_Jump, 0)

		// Store it and it's position in the same slice replacing the previous values
		jumpInstructions[i] = &(*cg.target)[len(*cg.target)-1]
		jumpInstructionPositions[i] = len(*cg.target)
	}

	// Calculate distance from end of each of if/elif body to the end. Assign it to the jump instructions.
	endPosition := len(*cg.target)
	for i, instruction := range jumpInstructions {
		updateJumpDistance(instruction, endPosition-jumpInstructionPositions[i], VM.IT_JumpEx)
	}

	// Assign distance to the end to the jump instruction in else block
	updateJumpDistance(jumpFromElse, endPosition-jumpFromElsePosition-1, VM.IT_JumpEx)
}
