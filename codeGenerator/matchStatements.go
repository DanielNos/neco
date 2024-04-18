package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateMatch(matchNode *parser.MatchNode) {
	// Create slice for jump instructions and their positions so their destination can be set later
	jumpInstructions := make([]*VM.Instruction, matchNode.CaseCount)
	jumpInstructionPositions := make([]int, matchNode.CaseCount)
	jumpIndex := 0

	// Generate matched expression
	cg.generateExpression(matchNode.Expression)

	// Generate case tests and jumps to their bodies
	for caseIndex, matchCase := range matchNode.Cases {
		caseNode := matchCase.Value.(*parser.CaseNode)
		for expressionIndex, expression := range caseNode.Expressions {
			// Duplicate matched expression and compare it to case expression
			if caseIndex < len(matchNode.Cases)-1 || expressionIndex < len(caseNode.Expressions)-1 {
				cg.addInstruction(VM.IT_DuplicateTop)
			}
			cg.generateExpression(expression)
			cg.addInstruction(VM.IT_Equal)

			// Generate conditional jump
			cg.addInstruction(VM.IT_JumpIfTrue, 0)

			// Store the jump instruction and it's position so destination position can be set later
			jumpInstructions[jumpIndex] = &(*cg.target)[len(*cg.target)-1]
			jumpInstructionPositions[jumpIndex] = len(*cg.target)
			jumpIndex++
		}
	}

	// Generate jump instruction that will jump over all case bodies
	cg.addInstruction(VM.IT_Jump, 0)
	// Store the jump instruction so it's destination can be set later
	jumpFromElse := &(*cg.target)[len(*cg.target)-1]
	jumpFromElsePosition := len(*cg.target)

	// Generate case bodies
	jumpIndex = 0
	for caseIndex, matchCase := range matchNode.Cases {
		caseNode := matchCase.Value.(*parser.CaseNode)

		// Set if's conditional jumps destination to next instruction
		intstructionIndex := len(*cg.target)
		for i := 0; i < len(caseNode.Expressions); i++ {
			updateJumpDistance(jumpInstructions[jumpIndex], intstructionIndex-jumpInstructionPositions[jumpIndex], VM.IT_JumpIfTrueEx)
			jumpIndex++
		}

		// Generate case body
		cg.generateNode(matchCase.Value.(*parser.CaseNode).Statement)

		// Generate jump instruction to the end of case bodies
		cg.addInstruction(VM.IT_Jump, 0)

		// Store it and it's position in the same slice replacing the previous values
		jumpInstructions[caseIndex] = &(*cg.target)[len(*cg.target)-1]
		jumpInstructionPositions[caseIndex] = len(*cg.target)
	}

	// Assign distance to the jump instruction for default case block
	if jumpFromElsePosition != len(*cg.target) {
		updateJumpDistance(jumpFromElse, len(*cg.target)-jumpFromElsePosition, VM.IT_JumpEx)
	}
	// Generate default body
	if matchNode.Default != nil {
		cg.generateNode(matchNode.Default)
	}

	// Calculate distance from end of each case body to the end. Assign it to the jump instructions.
	endPosition := len(*cg.target)
	for jumpIndex := 0; jumpIndex < len(matchNode.Cases); jumpIndex++ {
		updateJumpDistance(jumpInstructions[jumpIndex], endPosition-jumpInstructionPositions[jumpIndex], VM.IT_JumpEx)
	}

}
