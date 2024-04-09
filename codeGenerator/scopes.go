package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) enterScope(name *string) {
	if name == nil {
		cg.addInstruction(VM.IT_PushScopeUnnamed)
	} else {
		cg.addInstruction(VM.IT_PushScope, byte(cg.stringConstants[*name]))
	}
	cg.variableIdentifierCounters.Push(cg.variableIdentifierCounters.Top.Value)
	cg.variableIdentifiers.Push(map[string]uint8{})
}

func (cg *CodeGenerator) leaveScope() {
	cg.addInstruction(VM.IT_PopScope)
	cg.variableIdentifierCounters.Pop()
	cg.variableIdentifiers.Pop()
}

func (cg *CodeGenerator) generateStatements(scopeNode *parser.ScopeNode) {
	for _, node := range scopeNode.Statements {
		cg.generateNode(node)
	}
}

func (cg *CodeGenerator) generateScope(scopeNode *parser.ScopeNode, name *string) {
	cg.enterScope(name)
	cg.generateStatements(scopeNode)
	cg.leaveScope()
}
