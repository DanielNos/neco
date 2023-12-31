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

func (p *Parser) newErrorNoMessage(position *dataStructures.CodePos) {
	if p.ErrorCount + p.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

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

	// Enter or use current scope
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

		case lexer.TT_KW_var:
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

		// Return
		case lexer.TT_KW_return:
			returnPosition := p.consume().Position
			
			// Return value
			if p.peek().TokenType != lexer.TT_EndOfCommand {
				scope.statements = append(scope.statements, &Node{returnPosition, NT_Return, p.parseExpression(MINIMAL_PRECEDENCE)})
			// Return
			} else {
				scope.statements = append(scope.statements, &Node{returnPosition, NT_Return, nil})
			}
		
		// Skip EOCs
		case lexer.TT_EndOfCommand:
			p.consume()

		default:
			panic(fmt.Sprintf("Unexpected token \"%s\".", p.consume()))
		}
	}

	// Un-exited scope 
	if enterScope {
		p.scopeNodeStack.Pop()
		p.newError(opening, "Scope is missing a closing brace.")
	}

	return scope
}

func (p *Parser) parseIdentifier() *Node {
	identifier := p.consume()
	symbol := p.findSymbol(identifier.Value)

	// Assign to variable
	if p.peek().TokenType.IsAssignKeyword() {
		var expression *Node
		// Undeclared symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			expression, _ = p.parseAssign([]*lexer.Token{identifier}, []VariableType{{DT_NoType, false}})
		} else {
			// Assignment to function
			if symbol.symbolType == ST_Function {
				p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []VariableType{{DT_NoType, false}})
			// Assignment to variable
			} else {
				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []VariableType{symbol.value.(*VariableSymbol).variableType})
			}
		}
		return expression
	}

	// Assign to multiple variables
	if p.peek().TokenType == lexer.TT_DL_Comma {
		var identifiers = []*lexer.Token{identifier}
		var dataTypes = []VariableType{}

		// Check symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			dataTypes = append(dataTypes, VariableType{DT_NoType, false})
		} else if symbol.symbolType == ST_Function {
			p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
			dataTypes = append(dataTypes, VariableType{DT_NoType, false})
		} else {
			dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).variableType)
		}

		// Collect identifiers
		for p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()

			// Look up identifier and collect identifier
			symbol = p.findSymbol(p.peek().Value)
			identifiers = append(identifiers, p.consume())

			// Check symbol
			if symbol == nil {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Use of undeclared variable %s.", identifiers[len(identifiers) - 1]))
				dataTypes = append(dataTypes, VariableType{DT_NoType, false})
			} else if symbol.symbolType == ST_Function {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Can't assign to function %s.", identifiers[len(identifiers) - 1]))
				dataTypes = append(dataTypes, VariableType{DT_NoType, false})
			} else {
				dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).variableType)
			}
		}

		expression, _ := p.parseAssign(identifiers, dataTypes)
		return expression
	}

	// Function call
	// Undeclared function
	if symbol == nil {
		p.newError(identifier.Position, fmt.Sprintf("Use of undeclared function %s.", identifier.Value))
		return p.parseFunctionCall(symbol, p.consume())
	}

	// Declared function
	return p.parseFunctionCall(symbol, p.consume())
}

func (p *Parser) parseVariableDeclare() *Node {
	startPosition := p.peek().Position
	variableType := VariableType{TokenTypeToDataType[p.consume().TokenType], false}

	// Collect identifiers
	identifierTokens := []*lexer.Token{}
	identifiers := []string{}
	variableTypes := []VariableType{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		identifierTokens = append(identifierTokens, p.peek())
		identifiers = append(identifiers, p.peek().Value)
		variableTypes = append(variableTypes, variableType)

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
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		// var has to be assigned to
		if variableType.dataType == DT_NoType {
			startPosition.EndChar = p.peekPrevious().Position.EndChar
			p.newError(startPosition, "Variables declared using keyword var have to have an expression assigned to them, so a data type can be derived from it.")
		}
		p.consume()
	// Assign
	} else if p.peek().TokenType == lexer.TT_KW_Assign {
		// Push declare node
		node := declareNode
		p.appendScope(node)
		
		// Parse expression and collect type
		var expressionType VariableType
		declareNode, expressionType = p.parseAssign(identifierTokens, variableTypes)
		
		// Change variable type if no was provided
		if variableType.dataType == DT_NoType {
			variableType = expressionType
			node.value.(*VariableDeclareNode).variableType = expressionType
		}
	}

	// Insert symbols
	for _, id := range identifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{variableType, declareNode.nodeType == NT_Assign}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifierTokens []*lexer.Token, variableTypes []VariableType) (*Node, VariableType) {
	assignPosition := p.consume().Position
	expressionStart := p.peek().Position

	// Collect expression
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	// Get expression type
	expressionType := p.getExpressionType(expression)

	// Uncompatible data types
	expressionPosition := dataStructures.CodePos{File: expressionStart.File, Line: expressionStart.Line, StartChar: expressionStart.StartChar, EndChar: p.peekPrevious().Position.EndChar}

	// Print errors
	if expressionType.dataType != DT_NoType {
		for i, identifier := range identifierTokens {
			// Variable has a type and it's incompatible with expression
			if variableTypes[i].dataType != DT_NoType && !expressionType.Equals(variableTypes[i]) {
				p.newErrorNoMessage(&expressionPosition)
				logger.Error2CodePos(identifierTokens[i].Position, &expressionPosition, fmt.Sprintf("Can't assign expression of type %s to variable %s of type %s.", expressionType, identifier, variableTypes[i]))
			}
		}
	}

	var identifiers = []string{}
	
	for _, identifier := range identifierTokens {
		identifiers = append(identifiers, identifier.Value)
	}

	return &Node{assignPosition, NT_Assign, &AssignNode{identifiers, expression}}, expressionType
}
