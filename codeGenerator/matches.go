package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateMatch(matchNode *parser.MatchNode) {

	// Create slice for jump instructions and their positions so their destination can be set later
	jumpInstructions := make([]*VM.Instruction, len(matchNode.Cases))
	jumpInstructionPositions := make([]int, len(matchNode.Cases))

	// Generate matched expression
	cg.generateExpression(matchNode.Expression)

	// Generate else if conditions and jumps
	for i, matchCase := range matchNode.Cases {
		// Duplicate matched expression and compare it to case expression
		if i < len(matchNode.Cases)-1 {
			cg.addInstruction(VM.IT_DuplicateTop)
		}
		cg.generateExpression(matchCase.Value.(*parser.CaseNode).Expressions[0])
		cg.addInstruction(VM.IT_Equal)

		// Generate conditional jump
		cg.addInstruction(VM.IT_JumpIfTrue, 0)

		// Store the jump instruction and it's position so destination position can be set later
		jumpInstructions[i] = &(*cg.target)[len(*cg.target)-1]
		jumpInstructionPositions[i] = len(*cg.target)
	}

	// Generate default body
	if matchNode.Default != nil {
		cg.generateScope(matchNode.Default.Value.(*parser.ScopeNode), nil)
	}

	// Generate jump instruction that will jump over all elifs
	cg.addInstruction(VM.IT_Jump, 0)

	// Store the jump instruction so it's destination can  be set later
	jumpFromElse := &(*cg.target)[len(*cg.target)-1]

	// Store jump position
	jumpFromElsePosition := len(*cg.target)

	// Generate else if bodies
	for i, matchCase := range matchNode.Cases {
		// Set if's conditional jump destination to next instruction
		updateJumpDistance(jumpInstructions[i], len(*cg.target)-jumpInstructionPositions[i], VM.IT_JumpIfTrueEx)

		// Generate if's body
		cg.generateScope(matchCase.Value.(*parser.CaseNode).Statements.Value.(*parser.ScopeNode), nil)

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
	/*
		cg.generateExpression(matchNode.Expression)

		jumps := make([]*VM.Instruction, len(matchNode.Cases)-1)
		jumpPositions := make([]int, len(matchNode.Cases)-1)

		for i, matchCase := range matchNode.Cases[:len(matchNode.Cases)-2] {
			cg.addInstruction(VM.IT_DuplicateTop)
			cg.generateExpression(matchCase.Value.(*parser.MatchNode).Expression)

			cg.addInstruction(VM.IT_JumpIfTrue, 0)
			jumps[i] = &(*cg.target)[len(*cg.target)-1]
			jumpPositions[i] = len(*cg.target) - 1
		}

		cg.generateExpression(matchNode.Cases[len(matchNode.Cases)-2].Value.(*parser.MatchNode).Expression)
		cg.addInstruction(VM.IT_JumpIfTrue, 0)

		jumps[len(matchNode.Cases)-2] = &(*cg.target)[len(*cg.target)-1]
		jumpPositions[len(matchNode.Cases)-2] = len(*cg.target) - 1

		var jumpFromDefault *VM.Instruction = nil
		jumpFromDefaultPosition := 0

		if matchNode.Cases[len(matchNode.Cases)-1] != nil {
			for _, node := range matchNode.Cases[len(matchNode.Cases)-1].Value.(*parser.MatchNode).Cases {
				cg.generateNode(node)
			}
			cg.addInstruction(VM.IT_Jump, 0)
			jumpFromDefault = &(*cg.target)[len(*cg.target)-1]
			jumpFromDefaultPosition = len(*cg.target) - 1
		}

		for i, matchCase := range matchNode.Cases[:len(matchNode.Cases)-2] {
			jumps[i].InstructionValue[0] = byte(len(*cg.target) - jumpPositions[i])
			for _, node := range matchCase.Value.(*parser.MatchNode).Cases {
				cg.generateNode(node)
			}
		}

		jumps[len(matchNode.Cases)-2].InstructionValue[0] = byte(len(*cg.target) - jumpPositions[len(matchNode.Cases)-2])
		for _, node := range matchNode.Cases[len(matchNode.Cases)-2].Value.(*parser.MatchNode).Cases {
			cg.generateNode(node)
		}

		if jumpFromDefault != nil {
			jumpFromDefault.InstructionValue[0] = byte(len(*cg.target) - jumpFromDefaultPosition)
		}*/
}
