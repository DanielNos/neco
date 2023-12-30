package parser

import (
	"fmt"
	"neko/dataStructures"
	"neko/errors"
	"neko/lexer"
	"neko/logger"
	"os"
	"strings"
)

type Parser struct {
	tokens []*lexer.Token

	tokenIndex int

	scopeCounter int
	scopeNodeStack *dataStructures.Stack
	
	symbolTableStack *dataStructures.Stack

	ErrorCount uint
	totalErrorCount uint
}

func NewParser(tokens []*lexer.Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, dataStructures.NewStack(), dataStructures.NewStack(), 0, previousErrors}
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

func (p *Parser) newError(position *dataStructures.CodePos, message string) {
	if p.ErrorCount + p.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}
	
	logger.ErrorCodePos(position, message)
	p.ErrorCount++
	
	// Too many errors
	if p.ErrorCount + p.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.ERROR_SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.ErrorCount))
	}
}

func (p *Parser) insertSymbol(key string, symbol *Symbol) {
	p.symbolTableStack.Top.Value.(symbolTable)[key] = symbol
}

func (p *Parser) enterScope() {
	p.symbolTableStack.Push(symbolTable{})
	p.scopeNodeStack.Push(&ScopeNode{p.scopeCounter, []*Node{}})
	p.scopeCounter++
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
	p.enterScope()

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
	opening := p.consume().Position

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
				p.consume()
			// Root scope
			} else {
				p.newError(p.consume().Position, "Unexpected closing brace in root scope.")
			}

			return scope
		
		// Identifier
		case lexer.TT_Identifier:
			scope.statements = append(scope.statements, p.parseIdentifier())
			
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

func (p *Parser) parseIdentifier() *Node {
	symbol := p.findSymbol(p.peek().Value)

	// Undeclared symbol
	if symbol == nil {
		identifier := p.consume()

		if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared function %s.", identifier.Value))
			return p.parseFunctionCall(symbol, p.consume())
		} else {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			return p.parseAssign([]string{identifier.Value}, VariableType{DT_NoType, false})
		}

	} else {
		// Variable assignment
		if symbol.symbolType == ST_Variable {
			return p.parseAssign([]string{p.consume().Value}, symbol.value.(*VariableSymbol).variableType)
		} else if symbol.symbolType == ST_Function {
			return p.parseFunctionCall(symbol, p.consume())
		}
	}

	return nil
}

func (p *Parser) parseVariableDeclare() *Node {
	startPosition := p.peek().Position
	variableType := VariableType{TokenTypeToDataType[p.consume().TokenType], false}

	// Collect identifiers
	identifiers := []string{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		identifiers = append(identifiers, p.peek().Value)

		// Check if variable is redeclared
		symbol := p.getSymbol(p.peek().Value)

		if symbol != nil {
			p.newError(p.peek().Position, fmt.Sprintf("Variable %s is redeclared in this scope.", p.consume().Value))
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
	declareNode := &Node{startPosition, NT_VariableDeclare, &VariableDeclareNode{variableType, identifiers}}

	// End
	if p.peek().TokenType ==lexer. TT_EndOfCommand {
		p.consume()
	// Assign
	} else if p.peek().TokenType == lexer.TT_KW_Assign {
		p.appendScope(declareNode)
		declareNode = p.parseAssign(identifiers, variableType)
	}

	// Insert symbols
	for _, id := range identifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{variableType, declareNode.nodeType == NT_Assign}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifiers []string, variableType VariableType) *Node {
	assignPosition := p.consume().Position
	expressionStart := p.peek().Position

	// Collect expression
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	// Get expression type
	expressionType := p.getExpressionType(expression)

	// Uncompatible data types
	if expressionType.dataType != DT_NoType && !expressionType.Equals(variableType) {
		expressionPosition := dataStructures.CodePos{expressionStart.File, expressionStart.Line, expressionStart.StartChar, p.peekPrevious().Position.EndChar}
		p.newError(&expressionPosition, fmt.Sprintf("Can't assign expression of type %s to variable of type %s.", expressionType, variableType))
	}

	return &Node{assignPosition, NT_Assign, &AssignNode{identifiers, expression}}
}
