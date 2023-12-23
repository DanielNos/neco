package main

import (
	"fmt"
	"strings"
)

type Parser struct {
	tokens []*Token

	tokenIndex int

	scopeCounter int
	scopeNodeStack *Stack
	
	globalSymbolTable map[string]*Symbol

	errorCount uint
}

func NewParser(tokens []*Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, NewStack(), map[string]*Symbol{}, previousErrors}
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

func (p *Parser) newError(token *Token, message string) {
	p.errorCount++
	ErrorCodePos(token.position, message)

	// Too many errors
	if p.errorCount > MAX_ERROR_COUNT {
		Fatal(ERROR_SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.errorCount))
	}
}

func (p *Parser) insertSymbol(key string, symbol *Symbol) {
	if p.scopeNodeStack.size == 0 {
		p.globalSymbolTable[key] = symbol
	}
}

func (p *Parser) Parse() *Node {
	return p.parseModule()
}

func (p *Parser) collectGlobalSymbols() {
	for p.peek().tokenType != TT_EndOfFile {
		// Collect variable
		if p.peek().tokenType.IsVariableType() {
			p.consume()
		} else {
			p.consume()
		}
	}
}

func (p *Parser) parseModule() *Node {
	// Collect module path and name
	modulePath := p.consume().value
	pathParts := strings.Split(modulePath, "/")
	moduleName := pathParts[len(pathParts) - 1]

	if strings.Contains(moduleName, ".") { 
		moduleName = strings.Split(moduleName, ".")[0]
	}

	// Collect global symbols
	p.collectGlobalSymbols()
	p.tokenIndex = 0

	// Parse module
	p.consume()
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

		default:
			panic(fmt.Sprintf("Unexpected token \"%s\".", p.consume()))
		}
	}

	return scope
}

func (p *Parser) parseVariableDeclare() *Node {
	startPosition := p.peek().position
	dataType := TokenTypeToDataType[p.consume().tokenType]

	// Collect identifiers
	identifiers := []string{}
	identifiers = append(identifiers, p.consume().value)

	for p.peek().tokenType == TT_DL_Comma {
		p.consume()
		identifiers = append(identifiers, p.consume().value)
	}

	// Create node
	declareNode := &Node{startPosition, NT_VariableDeclare, &VariableDeclareNode{dataType, false, identifiers}}
	assigned := false

	// End
	if p.peek().tokenType == TT_EndOfCommand {
		p.consume()
	// Assign
	} else if p.peek().tokenType == TT_KW_Assign {
		p.appendScope(declareNode)
		declareNode = p.parseAssign(identifiers)
	}

	// Insert symbols
	for _, id := range identifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{dataType, false, assigned}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifiers []string) *Node {
	assign := p.consume()
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	return &Node{assign.position, NT_Assign, &AssignNode{identifiers, expression}}
}
