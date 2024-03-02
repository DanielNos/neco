package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateLoop(node *parser.Node) {
	// Record start position of loop
	startPosition := len(cg.instructions) - 1

	// Enter scope
	cg.enterScope()

	// Create break array and record loop scope
	cg.scopeBreaks.Push([]Break{})
	cg.loopScopeDepths.Push(cg.variableIdentifiers.Size)

	// Generate loop body
	cg.generateStatements(node.Value.(*parser.ScopeNode))

	// Leave loop scope
	cg.leaveScope()

	// Generate jump instruction back to start
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpBack, []byte{byte(len(cg.instructions) - startPosition)}})

	// Set destinations of break jumps
	instructionCount := len(cg.instructions)

	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		updateJumpDistance(b.instruction, instructionCount-b.instructionPosition, VM.IT_JumpEx)
	}
	cg.loopScopeDepths.Pop()
}

func (cg *CodeGenerator) generateForLoop(forLoop *parser.ForLoopNode) {

	// Enter scope
	cg.enterScope()

	// Create break array and record loop scope
	cg.scopeBreaks.Push([]Break{})
	cg.loopScopeDepths.Push(cg.variableIdentifiers.Size)

	// Generate init statement
	for _, node := range forLoop.InitStatement {
		cg.generateNode(node)
	}

	// Record start position of loop
	startPosition := len(cg.instructions) - 1

	// Generate loop body
	cg.generateStatements(forLoop.Body.Value.(*parser.ScopeNode))

	// Remove jump to start
	jumpInstruction := cg.instructions[len(cg.instructions)-1]
	cg.instructions = cg.instructions[:len(cg.instructions)-1]
	jumpPosition := len(cg.instructions)

	// Return adjusted jump instruction
	jumpInstruction.InstructionValue[0] += byte(len(cg.instructions) - jumpPosition)
	cg.instructions = append(cg.instructions, jumpInstruction)

	// Generate jump instruction back to start
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpBack, []byte{byte(len(cg.instructions) - startPosition)}})

	// Leave loop scope
	cg.leaveScope()

	// Set destinations of break jumps
	instructionCount := len(cg.instructions)

	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		updateJumpDistance(b.instruction, instructionCount-b.instructionPosition, VM.IT_JumpEx)
	}
	cg.loopScopeDepths.Pop()
}
