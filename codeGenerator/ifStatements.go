package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateIfStatement(ifStatement *parser.IfNode) {
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
		cg.generateStatements(ifStatement.ElseBody.Value.(*parser.ScopeNode))
	}

	// Generate jump instruction that will jump over all elifs
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})

	// Store the jump instruction so it's destination can  be set later
	jumpFromElse := &cg.instructions[len(cg.instructions)-1]

	// Store jump position
	jumpFromElsePosition := len(cg.instructions)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set elif's conditional jump destination to next instruction
		updateJumpDistance(jumpInstructions[i], len(cg.instructions)-jumpInstructionPositions[i], VM.IT_JumpIfTrueEx)

		// Generate elif's body
		cg.generateStatements(statement.Body.Value.(*parser.ScopeNode))

		// Generate jump instruction to the end of elif bodies
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})

		// Store it and it's position in the same array replacing the previous values
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

func updateJumpDistance(instruction *VM.Instruction, distance int, extendedInstructionType byte) {
	// If distance is larger than 255, change instruction type to extended jump
	if distance > MAX_UINT8 {
		instruction.InstructionType = extendedInstructionType
		instruction.InstructionValue = intTo2Bytes(distance)
	} else {
		instruction.InstructionValue[0] = byte(distance)
	}
}
