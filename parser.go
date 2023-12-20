package main

import (
	"github.com/golang-collections/collections/stack"
)

type Parser struct {
	tokens []*Token

	tokenIndex int

	scopeCounter int
	scopeNodeStack *stack.Stack

	errorCount uint
}

func NewParser(tokens []*Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, stack.New(), previousErrors}
}

func (p *Parser) peek() *Token {
	return p.tokens[p.tokenIndex]
}

func (p *Parser) consume() *Token {
	if p.tokenIndex + 1 < len(p.tokens) {
		p.tokenIndex++
	}
	return p.tokens[p.tokenIndex - 1]
}

func (p *Parser) currentScope() *ScopeNode {
	return p.scopeNodeStack.Peek().(*ScopeNode)
}

func (p *Parser) Parse() *Node {
	return p.parseModule()
}

func (p *Parser) parseModule() *Node {
	moduleName := p.consume().value

	scope := p.parseScope()

	var moduleNode NodeValue = &ModuleNode{moduleName, scope}
	module := &Node{p.peek().position, NT_Module, moduleNode}

	return module
}

func (p *Parser) parseScope() *ScopeNode {
	scope := &ScopeNode{p.scopeCounter, []*Node{}}
	p.scopeNodeStack.Push(scope)

	// Collect statements
	for p.peek().tokenType != TT_EndOfFile {
		switch p.peek().tokenType {
		// Variable declaration
		case TT_KW_bool, TT_KW_int, TT_KW_flt, TT_KW_str:
			scope.statements = append(scope.statements, p.parseVariableDeclare())

		case TT_EndOfCommand:
			p.consume()
		}

		p.consume()
	}

	return scope
}

func (p *Parser) parseVariableDeclare() *Node {
	dataType := p.consume()
	identifiers := []string{}

	// Collect identifiers
	identifiers = append(identifiers, p.consume().value)

	for p.peek().tokenType == TT_DL_Comma {
		p.consume()
		identifiers = append(identifiers, p.consume().value)
	}

	return &Node{dataType.position, NT_VariableDeclare, VariableDeclareNode{TokenTypeToDataType[dataType.tokenType], identifiers}}
}
