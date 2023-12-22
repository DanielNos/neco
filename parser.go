package main

import "strings"

type Parser struct {
	tokens []*Token

	tokenIndex int

	scopeCounter int
	scopeNodeStack *Stack

	errorCount uint
}

func NewParser(tokens []*Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, NewStack(), previousErrors}
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

func (p *Parser) appendScope(node *Node) {
	p.scopeNodeStack.top.value.(*ScopeNode).statements = append(p.scopeNodeStack.top.value.(*ScopeNode).statements, node)
}

func (p *Parser) Parse() *Node {
	return p.parseModule()
}

func (p *Parser) parseModule() *Node {
	// Collect module path and name
	modulePath := p.consume().value
	pathParts := strings.Split(modulePath, "/")
	moduleName := pathParts[len(pathParts)-1]

	if strings.Contains(moduleName, ".") { 
		moduleName = strings.Split(moduleName, ".")[0]
	}

	// Parse module
	scope := p.parseScope()

	// Create node
	var moduleNode NodeValue = &ModuleNode{modulePath, moduleName, scope}
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

	declareNode := &Node{dataType.position, NT_VariableDeclare, &VariableDeclareNode{TokenTypeToDataType[dataType.tokenType], identifiers}}

	if p.peek().tokenType == TT_EndOfCommand {
		p.consume()
	} else if p.peek().tokenType == TT_KW_Assign {
		p.appendScope(declareNode)
		return p.parseAssign(identifiers)
	}

	return declareNode
}

func (p *Parser) parseAssign(identifiers []string) *Node {
	assign := p.consume()
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	return &Node{assign.position, NT_Assign, &AssignNode{identifiers, expression}}
}
