package codeGenerator

import (
	"github.com/DanielNos/NeCo/parser"
	VM "github.com/DanielNos/NeCo/virtualMachine"
)

func (cg *CodeGenerator) pushScope(scopeType ScopeType) {
	varIdCount := uint8(0)

	if cg.scopes.Top != nil {
		varIdCount = cg.scopes.Top.Value.(*Scope).variableIdentifierCounter
	}

	cg.scopes.Push(&Scope{
		scopeType,
		varIdCount,
		map[string]uint8{},
	})
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
