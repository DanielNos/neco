package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) pushScope(scopeType ScopeType) {
	cg.scopes.Push(&Scope{scopeType, uint8(0), map[string]uint8{}})
}

func (cg *CodeGenerator) enterScope(name *string) {
	if name == nil {
		cg.addInstruction(VM.IT_PushScopeUnnamed)
		cg.pushScope(ST_Unnamed)
	} else {
		cg.addInstruction(VM.IT_PushScope, byte(cg.stringConstants[*name]))
		cg.pushScope(ST_Function)
	}
}

func (cg *CodeGenerator) leaveScope() {
	cg.addInstruction(VM.IT_PopScope)
	cg.scopes.Pop()
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
