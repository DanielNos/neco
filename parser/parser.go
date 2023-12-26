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
	
	symbolTableStack *dataStructures.Stack

	ErrorCount uint
}

func NewParser(tokens []*lexer.Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, dataStructures.NewStack(), dataStructures.NewStack(), previousErrors}
}

func (p *Parser) peek() *lexer.Token {
	return p.tokens[p.tokenIndex]
}

func (p *Parser) peekPrevious() *lexer.Token {
	if p.tokenIndex > 0 {
		return p.tokens[p.tokenIndex - 1]
	}
	return p.tokens[0]
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
	p.symbolTableStack.Top.Value.(map[string]*Symbol)[key] = symbol
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

	// Enter global scope
	p.symbolTableStack.Push(symbolTable{})
	p.scopeNodeStack.Push(&ScopeNode{p.scopeCounter, []*Node{}})

	// Collect global symbols
	p.collectGlobalSymbols()
	p.tokenIndex = 0

	// Parse module
	scope := p.parseScope(false)

	// Create node
	var moduleNode NodeValue = &ModuleNode{modulePath, moduleName, scope}
	module := &Node{p.peek().Position, NT_Module, moduleNode}

	return module
}

func (p *Parser) parseScope(enterScope bool) *ScopeNode {
	// Consume opening brace
	opening := p.consume()

	var scope *ScopeNode
	
	if enterScope {
		scope = &ScopeNode{p.scopeCounter, []*Node{}}
		p.scopeNodeStack.Push(scope)
		p.scopeCounter++
	} else {
		scope = p.scopeNodeStack.Top.Value.(*ScopeNode)
	}

	// Collect statements
	for p.peek().TokenType != lexer.TT_EndOfFile {
		switch p.peek().TokenType {

		// Variable declaration
		case lexer.TT_KW_bool, lexer.TT_KW_int, lexer.TT_KW_flt, lexer.TT_KW_str:
			scope.statements = append(scope.statements, p.parseVariableDeclare())

		// Function declaration
		case lexer.TT_KW_fun:
			scope.statements = append(scope.statements, p.parseFunctionDeclare())

		// Leave scope
		case lexer.TT_DL_BraceClose:
			// Pop scope
			if p.scopeNodeStack.Size > 1 {
				if enterScope {
					p.scopeNodeStack.Pop()
					p.symbolTableStack.Pop()
				}
			// Root scope
			} else {
				p.newError(p.consume(), "Unexpected closing brace in root scope.")
			}
			return scope
			
		case lexer.TT_EndOfCommand:
			p.consume()

		default:
			panic(fmt.Sprintf("Unexpected token \"%s\".", p.consume()))
		}
	}

	if enterScope {
		p.scopeNodeStack.Pop()
		p.newError(opening, "Scope is missing a closing brace.")
	}

	return scope
}

func (p *Parser) parseVariableDeclare() *Node {
	startPosition := p.peek().Position
	dataType := TokenTypeToDataType[p.consume().TokenType]

	// Collect identifiers
	identifiers := []string{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		identifiers = append(identifiers, p.peek().Value)

		// Check if variable is redeclared
		_, exists := p.symbolTableStack.Top.Value.(symbolTable)[p.peek().Value]

		if exists {
			p.newError(p.peek(), fmt.Sprintf("Variable %s is redeclared in this scope.", p.consume().Value))
		} else {
			p.consume()
		}

		// More identifiers
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
		} else {
			break
		}
	}

	// Create node
	declareNode := &Node{startPosition, NT_VariableDeclare, &VariableDeclareNode{dataType, false, identifiers}}

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
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{dataType, false, declareNode.nodeType == NT_Assign}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifiers []string) *Node {
	assign := p.consume()
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	return &Node{assign.Position, NT_Assign, &AssignNode{identifiers, expression}}
}
