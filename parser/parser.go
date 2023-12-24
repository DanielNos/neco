package parser

import (
	"fmt"
	"neko/dataStructures"
	"neko/errors"
	"neko/lexer"
	"neko/logger"
	"strings"
)

type Parser struct {
	tokens []*lexer.Token

	tokenIndex int

	scopeCounter int
	scopeNodeStack *dataStructures.Stack
	
	globalSymbolTable map[string]*Symbol

	ErrorCount uint
}

func NewParser(tokens []*lexer.Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, dataStructures.NewStack(), map[string]*Symbol{}, previousErrors}
}

func (p *Parser) peek() *lexer.Token {
	return p.tokens[p.tokenIndex]
}

func (p *Parser) consume() *lexer.Token {
	if p.tokenIndex + 1 < len(p.tokens) {
		p.tokenIndex++
	}
	return p.tokens[p.tokenIndex - 1]
}

func (p *Parser) appendScope(node *Node) {
	p.scopeNodeStack.Top.Value.(*ScopeNode).statements = append(p.scopeNodeStack.Top.Value.(*ScopeNode).statements, node)
}

func (p *Parser) newError(token *lexer.Token, message string) {
	p.ErrorCount++
	logger.ErrorCodePos(token.Position, message)

	// Too many errors
	if p.ErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.ERROR_SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.ErrorCount))
	}
}

func (p *Parser) insertSymbol(key string, symbol *Symbol) {
	if p.scopeNodeStack.Size == 0 {
		p.globalSymbolTable[key] = symbol
	}
}

func (p *Parser) Parse() *Node {
	return p.parseModule()
}

func (p *Parser) collectGlobalSymbols() {
	for p.peek().TokenType != lexer.TT_EndOfFile {
		// Collect variable
		if p.peek().TokenType.IsVariableType() {
			p.consume()
		} else {
			p.consume()
		}
	}
}

func (p *Parser) parseModule() *Node {
	// Collect module path and name
	modulePath := p.consume().Value
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
	module := &Node{p.peek().Position, NT_Module, moduleNode}

	return module
}

func (p *Parser) parseScope() *ScopeNode {
	scope := &ScopeNode{p.scopeCounter, []*Node{}}
	p.scopeNodeStack.Push(scope)

	// Collect statements
	for p.peek().TokenType != lexer.TT_EndOfFile {
		switch p.peek().TokenType {
		// Variable declaration
		case lexer.TT_KW_bool, lexer.TT_KW_int, lexer.TT_KW_flt, lexer.TT_KW_str:
			scope.statements = append(scope.statements, p.parseVariableDeclare())

		case lexer.TT_EndOfCommand:
			p.consume()

		default:
			panic(fmt.Sprintf("Unexpected token \"%s\".", p.consume()))
		}
	}

	return scope
}

func (p *Parser) parseVariableDeclare() *Node {
	startPosition := p.peek().Position
	dataType := TokenTypeToDataType[p.consume().TokenType]

	// Collect identifiers
	identifiers := []string{}
	identifiers = append(identifiers, p.consume().Value)

	for p.peek().TokenType == lexer.TT_DL_Comma {
		p.consume()
		identifiers = append(identifiers, p.consume().Value)
	}

	// Create node
	declareNode := &Node{startPosition, NT_VariableDeclare, &VariableDeclareNode{dataType, false, identifiers}}
	assigned := false

	// End
	if p.peek().TokenType ==lexer. TT_EndOfCommand {
		p.consume()
	// Assign
	} else if p.peek().TokenType == lexer.TT_KW_Assign {
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

	return &Node{assign.Position, NT_Assign, &AssignNode{identifiers, expression}}
}
