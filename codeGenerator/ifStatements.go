package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateIfStatement(ifStatement *parser.IfNode) {
	// Create slice for jump instructions and their positions so their destination can be set later
	jumpInstructions := make([]*VM.Instruction, len(ifStatement.IfStatements))
	jumpInstructionPositions := make([]int, len(ifStatement.IfStatements))

	// Generate else if conditions and jumps
	for i, statement := range ifStatement.IfStatements {
		// Generate condition expression
		cg.generateExpression(statement.Condition, true)
		// Generate conditional jump
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpIfTrue, []byte{0}})

		// Store the jump instruction and it's position so destination position can be set later
		jumpInstructions[i] = &cg.instructions[len(cg.instructions)-1]
		jumpInstructionPositions[i] = len(cg.instructions)
	}

	// Generate else body
	if ifStatement.ElseBody != nil {
		cg.generateScope(ifStatement.ElseBody.Value.(*parser.ScopeNode))
	}

	// Generate jump instruction that will jump over all elifs
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})

	// Store the jump instruction so it's destination can  be set later
	jumpFromElse := &cg.instructions[len(cg.instructions)-1]

	// Store jump position
	jumpFromElsePosition := len(cg.instructions)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set if's conditional jump destination to next instruction
		updateJumpDistance(jumpInstructions[i], len(cg.instructions)-jumpInstructionPositions[i], VM.IT_JumpIfTrueEx)

		// Generate if's body
		cg.generateScope(statement.Body.Value.(*parser.ScopeNode))

		// Generate jump instruction to the end of if bodies
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})

		// Store it and it's position in the same slice replacing the previous values
		jumpInstructions[i] = &cg.instructions[len(cg.instructions)-1]
		jumpInstructionPositions[i] = len(cg.instructions)
	}

	// Calculate distance from end of each of if/elif body to the end. Assign it to the jump instructions.
	endPosition := len(cg.instructions)
	for i, instruction := range jumpInstructions {
		updateJumpDistance(instruction, endPosition-jumpInstructionPositions[i], VM.IT_JumpEx)
	}

	// Assign distance to the end to the jump instruction in else block
	updateJumpDistance(jumpFromElse, endPosition-jumpFromElsePosition, VM.IT_JumpEx)
	jumpFromElse.InstructionValue[0] = byte(endPosition - jumpFromElsePosition)
}
