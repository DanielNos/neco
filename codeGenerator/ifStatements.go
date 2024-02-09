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
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpIfTrue, ARGS_ZERO})

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
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, ARGS_ZERO})
		// Store the jump instruction so it's destination can  be set later
		jumpFromElse = &cg.instructions[len(cg.instructions)-1]
	}
	// Store jump position
	jumpFromElsePosition := len(cg.instructions)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set elif's conditional jump destination to next instruction
		jumpInstructions[i].InstructionValue[0] = byte(len(cg.instructions) - jumpInstructionPositions[i])
		// Generate elif's body
		cg.generateStatements(statement.Body.Value.(*parser.ScopeNode))

		// Generate jump instruction to the end of elif bodies
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, ARGS_ZERO})
		// Store it and it's position in the same array replacing the previous values
		jumpInstructions[i] = &cg.instructions[len(cg.instructions)-1]
		jumpInstructionPositions[i] = len(cg.instructions)
	}

	// Calculate distance from end of each of if/elif body to the end. Assign it to the jump instructions.
	endPosition := len(cg.instructions)
	for i, instruction := range jumpInstructions {
		instruction.InstructionValue[0] = byte(endPosition - jumpInstructionPositions[i])
	}

	// Assign distance to the end to the jump instruction in else block
	if jumpFromElse != nil {
		jumpFromElse.InstructionValue[0] = byte(endPosition - jumpFromElsePosition)
	}
}
