package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateLoop(node *parser.Node) {
	// Enter scope and create
	cg.enterScope()
	// Create break array and record loop scope
	cg.scopeBreaks.Push([]Break{})
	cg.loopScopeDepths.Push(cg.variableIdentifiers.Size)

	// Record start position of loops
	startPosition := len(cg.instructions)

	// Generate loop body
	cg.generateStatements(node.Value.(*parser.ScopeNode))

	// Generate jump instruction back to start
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpBack, []byte{byte(len(cg.instructions) - startPosition)}})

	// Set destinations of break jumps
	instructionCount := len(cg.instructions)

	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		updateJumpDistance(b.instruction, instructionCount-b.instructionPosition, VM.IT_JumpIfTrueEx)
	}
	cg.loopScopeDepths.Pop()

	// Leave loop scope
	cg.leaveScope()
}
