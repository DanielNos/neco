package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateLoop(node *parser.Node) {
	// Enter scope and create an array for breaks
	cg.enterScope()
	cg.scopeBreaks.Push([]Break{})

	// Record start position of loops
	startPosition := len(cg.instructions)

	// Generate loop body
	cg.generateStatements(node.Value.(*parser.ScopeNode))

	// Generate jump instruction back to start
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpBack, []byte{byte(len(cg.instructions) - startPosition)}})

	// Set destinations of break jumps
	distance := 0
	instructionCount := len(cg.instructions)
	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		distance = instructionCount - b.instructionPosition

		// If distance is larger than 255, change instruction type to extended jump
		if distance > MAX_UINT8 {
			b.instruction.InstructionType = VM.IT_JumpIfTrueEx
			b.instruction.InstructionValue = intTo2Bytes(distance)
		} else {
			b.instruction.InstructionValue[0] = byte(distance)
		}
	}

	// Leave loop scope
	cg.leaveScope()
}
