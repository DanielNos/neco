package codeGenerator

import (
	"github.com/DanielNos/neco/parser"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func (cg *CodeGenerator) generateLoop(node *parser.Node) {
	// Record start position of loop
	startPosition := len(*cg.target) - 1

	// Enter scope
	cg.enterScope(nil)

	// Create break array and record loop scope
	cg.scopeBreaks.Push([]Break{})
	cg.loopScopeDepths.Push(cg.scopes.Size)

	// Generate loop body
	cg.generateStatements(node.Value.(*parser.ScopeNode))

	// Leave loop scope
	cg.leaveScope()

	// Generate line offset if line changed
	if cg.line < node.Position.EndLine {
		cg.addInstruction(VM.IT_LineOffset, byte(node.Position.EndLine-cg.line))
		cg.line = node.Position.EndLine
	}

	// Generate jump instruction back to start
	cg.addInstruction(VM.IT_JumpBack, byte(len(*cg.target)-startPosition))

	// Set destinations of break jumps
	instructionCount := len(*cg.target)

	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		updateJumpDistance(b.instruction, instructionCount-b.instructionPosition, VM.IT_JumpEx)
	}
	cg.loopScopeDepths.Pop()
}

func (cg *CodeGenerator) generateForLoop(node *parser.Node) {
	forLoop := node.Value.(*parser.ForLoopNode)

	// Enter scope
	cg.enterScope(nil)

	// Create break array and record loop scope
	cg.scopeBreaks.Push([]Break{})
	cg.loopScopeDepths.Push(cg.scopes.Size)

	// Generate init statement
	for _, node := range forLoop.InitStatement {
		cg.generateNode(node)
	}

	// Record start position of loop
	startPosition := len(*cg.target) - 1

	// Generate loop body
	cg.generateStatements(forLoop.Body.Value.(*parser.ScopeNode))

	// Generate line offset if line changed
	if cg.line < node.Position.EndLine {
		cg.addInstruction(VM.IT_LineOffset, byte(node.Position.EndLine-cg.line))
		cg.line = node.Position.EndLine
	}

	// Remove jump to start
	jumpInstruction := (*cg.target)[len(*cg.target)-1]
	*cg.target = (*cg.target)[:len(*cg.target)-1]
	jumpPosition := len(*cg.target)

	// Generate return adjusted jump instruction
	jumpInstruction.InstructionValue[0] += byte(len(*cg.target) - jumpPosition)
	*cg.target = append(*cg.target, jumpInstruction)

	// Generate jump instruction back to start
	cg.addInstruction(VM.IT_JumpBack, byte(len(*cg.target)-startPosition))

	// Leave loop scope
	cg.leaveScope()

	// Set destinations of break jumps
	instructionCount := len(*cg.target)

	for _, b := range cg.scopeBreaks.Pop().([]Break) {
		updateJumpDistance(b.instruction, instructionCount-b.instructionPosition, VM.IT_JumpEx)
	}
	cg.loopScopeDepths.Pop()
}
