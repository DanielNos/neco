package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

const MAX_UINT8 = 255

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
	var jumpFromElse *VM.Instruction = nil
	if ifStatement.ElseBody != nil {
		// Generate body
		cg.generateStatements(ifStatement.ElseBody.Value.(*parser.ScopeNode))

		// Generate jump instruction that will jump over all elifs
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})
		// Store the jump instruction so it's destination can  be set later
		jumpFromElse = &cg.instructions[len(cg.instructions)-1]
	}
	// Store jump position
	jumpFromElsePosition := len(cg.instructions)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set elif's conditional jump destination to next instruction
		distance := len(cg.instructions) - jumpInstructionPositions[i]

		// If distance is larger than 255, change instruction type to extended jump
		if distance > MAX_UINT8 {
			jumpInstructions[i].InstructionType = VM.IT_JumpIfTrueEx
			jumpInstructions[i].InstructionValue = intTo2Bytes(distance)
		} else {
			jumpInstructions[i].InstructionValue[0] = byte(distance)
		}

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
		distance := endPosition - jumpInstructionPositions[i]

		// If distance is larger than 255, change instruction type to extended jump
		if distance > MAX_UINT8 {
			instruction.InstructionType = VM.IT_JumpEx
			instruction.InstructionValue = intTo2Bytes(distance)
		} else {
			instruction.InstructionValue[0] = byte(distance)
		}
	}

	// Assign distance to the end to the jump instruction in else block
	if jumpFromElse != nil {
		distance := endPosition - jumpFromElsePosition

		// If distance is larger than 255, change instruction type to extended
		if distance > MAX_UINT8 {
			jumpFromElse.InstructionType = VM.IT_JumpEx
			jumpFromElse.InstructionValue = intTo2Bytes(distance)
		} else {
			jumpFromElse.InstructionValue[0] = byte(distance)
		}
		jumpFromElse.InstructionValue[0] = byte(endPosition - jumpFromElsePosition)
	}
}
