package parser

import (
	"fmt"
	"neco/dataStructures"
	"neco/errors"
	"neco/lexer"
	"neco/logger"
	"os"
	"strings"
)

type Parser struct {
	tokens []*lexer.Token

	tokenIndex int

	scopeCounter   int
	scopeNodeStack *dataStructures.Stack

	stack_symbolTablestack *dataStructures.Stack

	functions     []*FunctionSymbol
	functionIndex int

	ErrorCount      uint
	totalErrorCount uint

	IntConstants    map[int64]int
	FloatConstants  map[float64]int
	StringConstants map[string]int
}

func NewParser(tokens []*lexer.Token, previousErrors uint) Parser {
	return Parser{tokens, 0, 0, dataStructures.NewStack(), dataStructures.NewStack(), []*FunctionSymbol{}, 0, 0, previousErrors, map[int64]int{}, map[float64]int{}, map[string]int{}}
}

func (p *Parser) peek() *lexer.Token {
	return p.tokens[p.tokenIndex]
}

func (p *Parser) peekNext() *lexer.Token {
	if p.tokenIndex+1 < len(p.tokens) {
		return p.tokens[p.tokenIndex+1]
	}
	return p.tokens[p.tokenIndex]
}

func (p *Parser) peekPrevious() *lexer.Token {
	if p.tokenIndex > 0 {
		return p.tokens[p.tokenIndex-1]
	}
	return p.tokens[0]
}

func (p *Parser) consume() *lexer.Token {
	if p.tokenIndex+1 < len(p.tokens) {
		p.tokenIndex++
	}
	return p.tokens[p.tokenIndex-1]
}

func (p *Parser) appendScope(node *Node) {
	p.scopeNodeStack.Top.Value.(*ScopeNode).Statements = append(p.scopeNodeStack.Top.Value.(*ScopeNode).Statements, node)
}

func (p *Parser) newError(position *dataStructures.CodePos, message string) {
	if p.ErrorCount+p.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	logger.ErrorCodePos(position, message)
	p.ErrorCount++

	// Too many errors
	if p.ErrorCount+p.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.ErrorCount))
	}
}

func (p *Parser) newErrorNoMessage(position *dataStructures.CodePos) {
	if p.ErrorCount+p.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	p.ErrorCount++

	// Too many errors
	if p.ErrorCount+p.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.ErrorCount))
	}
}

func (p *Parser) insertSymbol(key string, symbol *Symbol) {
	p.stack_symbolTablestack.Top.Value.(symbolTable)[key] = symbol
}

func (p *Parser) enterScope() {
	p.stack_symbolTablestack.Push(symbolTable{})
	p.scopeNodeStack.Push(&ScopeNode{p.scopeCounter, []*Node{}})
	p.scopeCounter++
}

func (p *Parser) leaveScope() {
	p.scopeNodeStack.Pop()
	p.stack_symbolTablestack.Pop()
}

func (p *Parser) Parse() *Node {
	return p.parseModule()
}

func (p *Parser) parseModule() *Node {
	// Collect module path and name
	modulePath := p.consume().Value
	pathParts := strings.Split(modulePath, "/")
	moduleName := pathParts[len(pathParts)-1]

	if strings.Contains(moduleName, ".") {
		moduleName = strings.Split(moduleName, ".")[0]
	}

	// Enter global scope
	p.enterScope()

	// Insert built-in functions
	p.insertBuiltInFunctions()

	// Collect function headers
	p.collectFunctionHeaders()
	p.tokenIndex = 1

	// Parse module
	scopeNode := p.parseScope(false, false)

	// Create node
	var moduleNode NodeValue = &ModuleNode{modulePath, moduleName, scopeNode.(*ScopeNode)}
	module := &Node{p.peek().Position, NT_Module, moduleNode}

	// No entry function
	if p.getGlobalSymbol("entry") == nil {
		logger.Warning("The entry() function wasn't found. The compiled program won't be executable by itself.")
	}

	// Check if all functions were called
	for identifier, symbol := range p.stack_symbolTablestack.Bottom.Value.(symbolTable) {
		// Try to find function bucket symbol
		if symbol.symbolType == ST_FunctionBucket {
			// Check if every function in the bucket was ever called
			for _, functionSymbol := range symbol.value.(symbolTable) {
				if !functionSymbol.value.(*FunctionSymbol).everCalled {
					logger.Warning(fmt.Sprintf("Function %s was never called.", identifier))
				}
			}
		}

	}

	return module
}

func (p *Parser) collectFunctionHeaders() {
	for p.peek().TokenType != lexer.TT_EndOfFile {
		if p.peek().TokenType == lexer.TT_KW_fun {
			p.consume()
			p.parseFunctionHeader()
		} else {
			p.consume()
		}
	}
}

func (p *Parser) parseScope(enterScope, packInNode bool) interface{} {
	// Consume opening brace
	opening := p.peek().Position

	if p.peek().TokenType == lexer.TT_DL_BraceOpen {
		p.consume()
	}

	// Enter or use current scope
	var scope *ScopeNode

	if enterScope {
		p.enterScope()
	}
	scope = p.scopeNodeStack.Top.Value.(*ScopeNode)

	// Collect statements
	for p.peek().TokenType != lexer.TT_EndOfFile {
		statement := p.parseStatement(enterScope)

		if statement == nil {
			if packInNode {
				return &Node{opening, NT_Scope, scope}
			} else {
				return scope
			}
		}

		scope.Statements = append(scope.Statements, statement)
	}

	// Un-exited scope
	if enterScope {
		p.scopeNodeStack.Pop()
		p.newError(opening, "Scope is missing a closing brace.")
	}

	if packInNode {
		return &Node{opening, NT_Scope, scope}
	} else {
		return scope
	}
}

func (p *Parser) parseStatement(enteredScope bool) *Node {
	switch p.peek().TokenType {

	// Variable declaration
	case lexer.TT_KW_bool, lexer.TT_KW_int, lexer.TT_KW_flt, lexer.TT_KW_str, lexer.TT_KW_list:
		return p.parseVariableDeclare(false)

	case lexer.TT_KW_var:
		return p.parseVariableDeclare(false)

	// Constant variable
	case lexer.TT_KW_const:
		p.consume()
		return p.parseVariableDeclare(true)

	// Function declaration
	case lexer.TT_KW_fun:
		return p.parseFunctionDeclare()

	// Leave scope
	case lexer.TT_DL_BraceClose:
		// Pop scope
		if p.scopeNodeStack.Size > 1 {
			if enteredScope {
				p.scopeNodeStack.Pop()
				p.stack_symbolTablestack.Pop()
			}
			p.consume()
			// Root scope
		} else {
			p.newError(p.consume().Position, "Unexpected closing brace in root scope.")
		}

		return nil

	// Identifier
	case lexer.TT_Identifier:
		return p.parseIdentifier()

	// Return
	case lexer.TT_KW_return:
		returnPosition := p.consume().Position

		// Return value
		if p.peek().TokenType != lexer.TT_EndOfCommand {
			return &Node{returnPosition, NT_Return, p.parseExpressionRoot()}
			// Return
		} else {
			return &Node{returnPosition, NT_Return, nil}
		}

	// If statement
	case lexer.TT_KW_if:
		return p.parseIfStatement()

	// Loop
	case lexer.TT_KW_loop:
		return p.parseLoop()

	// While
	case lexer.TT_KW_while:
		return p.parseWhile()

	// For
	case lexer.TT_KW_for:
		return p.parseFor()

	case lexer.TT_EndOfFile:
		return nil

	// Scope
	case lexer.TT_DL_BraceOpen:
		return p.parseScope(true, true).(*Node)

	// Break
	case lexer.TT_KW_break:
		return &Node{p.consume().Position, NT_Break, nil}

	// Skip EOCs
	case lexer.TT_EndOfCommand:
		p.consume()
		return p.parseStatement(enteredScope)
	}

	panic(fmt.Sprintf("%v Unexpected token %s \"%s\".", p.peek().Position, p.peek().TokenType, p.consume()))
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
			expression, _ = p.parseAssign([]*lexer.Token{identifier}, []DataType{{DT_NoType, nil}})
		} else {
			// Assignment to function
			if symbol.symbolType == ST_Function {
				p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []DataType{{DT_NoType, nil}})
				// Assignment to variable
			} else {
				// Can't assign to constants
				if symbol.value.(*VariableSymbol).isConstant {
					p.newError(p.peek().Position, fmt.Sprintf("Can't assign to constant variable %s.", identifier.Value))
				}

				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []DataType{symbol.value.(*VariableSymbol).VariableType})

				symbol.value.(*VariableSymbol).isInitialized = true
			}
		}
		return expression
	}

	// Assign to list at index
	if p.peek().TokenType == lexer.TT_DL_BracketOpen {
		var listType DataType = DataType{DT_NoType, nil}

		// Undeclared symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
		} else {
			// Not a variable
			if symbol.symbolType != ST_Variable {
				p.newError(identifier.Position, fmt.Sprintf("Can't assign to %s. It is not a variable.", identifier.Value))
				// Collect list type
			} else {
				listType = symbol.value.(*VariableSymbol).VariableType
			}
		}

		// Collect index expression
		p.consume()
		indexExpression := p.parseExpressionRoot()
		p.consume()

		// Index must be int
		if !p.GetExpressionType(indexExpression).Equals(DataType{DT_Int, nil}) {
			p.newError(getExpressionPosition(indexExpression, indexExpression.Position.StartChar, indexExpression.Position.EndChar), "Index expression has to be int.")
		}

		assignPosition := p.consume().Position

		// Collect assigned expression
		assignedExpression := p.parseExpressionRoot()
		assignedType := p.GetExpressionType(assignedExpression)

		// Check if assigned expression has the correct type
		if listType.DType != DT_NoType && !assignedType.Equals(listType.SubType.(DataType)) {
			expressionPosition := getExpressionPosition(assignedExpression, assignedExpression.Position.StartChar, assignedExpression.Position.EndChar)
			p.newError(expressionPosition, fmt.Sprintf("Can't assign expression of type %s to %s.", assignedType, listType))
		}

		return &Node{assignPosition, NT_ListAssign, &ListAssignNode{identifier.Value, symbol.value.(*VariableSymbol), indexExpression, assignedExpression}}
	}

	// Assign to multiple variables
	if p.peek().TokenType == lexer.TT_DL_Comma {
		var dataTypes = []DataType{}

		// Check symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			dataTypes = append(dataTypes, DataType{DT_NoType, nil})
		} else if symbol.symbolType == ST_Function {
			p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
			dataTypes = append(dataTypes, DataType{DT_NoType, nil})
		} else {
			dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).VariableType)
		}

		// Collect identifiers
		var identifiers = []*lexer.Token{identifier}
		symbols := []*Symbol{}

		for p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()

			// Look up identifier and collect it
			symbol = p.findSymbol(p.peek().Value)
			identifiers = append(identifiers, p.consume())

			// Check symbol
			if symbol == nil {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Use of undeclared variable %s.", identifiers[len(identifiers)-1]))
				dataTypes = append(dataTypes, DataType{DT_NoType, nil})
			} else if symbol.symbolType == ST_Function {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Can't assign to function %s.", identifiers[len(identifiers)-1]))
				dataTypes = append(dataTypes, DataType{DT_NoType, nil})
			} else {
				dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).VariableType)
			}

			symbols = append(symbols, symbol)
		}

		expression, _ := p.parseAssign(identifiers, dataTypes)

		// Set symbols as initialized
		for _, symbol := range symbols {
			symbol.value.(*VariableSymbol).isInitialized = true
		}

		return expression
	}

	// Function call
	// Undeclared function
	if symbol == nil {
		p.newError(identifier.Position, fmt.Sprintf("Use of undeclared function %s.", identifier.Value))
		return p.parseFunctionCall(symbol, identifier)
	}

	// Declared function
	return p.parseFunctionCall(symbol, identifier)
}

func (p *Parser) parseVariableDeclare(constant bool) *Node {
	startPosition := p.peek().Position

	// Collect data type
	variableType := DataType{TokenTypeToDataType[p.consume().TokenType], nil}

	if p.peek().TokenType == lexer.TT_OP_Lower {
		p.consume()
		variableType.SubType = DataType{TokenTypeToDataType[p.consume().TokenType], nil}
		p.consume()
	}

	// Collect identifiers
	identifierTokens := []*lexer.Token{}
	identifiers := []string{}
	variableTypes := []DataType{}

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
	declareNode := &Node{startPosition, NT_VariableDeclare, &VariableDeclareNode{variableType, constant, identifiers}}

	// End
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		// var has to be assigned to
		if variableType.DType == DT_NoType {
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
		var expressionType DataType
		declareNode, expressionType = p.parseAssign(identifierTokens, variableTypes)

		// Change variable type if no was provided
		if variableType.DType == DT_NoType {
			variableType = expressionType
			node.Value.(*VariableDeclareNode).DataType = expressionType
		}
	}

	// Insert symbols
	for _, id := range identifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{variableType, declareNode.NodeType == NT_Assign, constant}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifierTokens []*lexer.Token, variableTypes []DataType) (*Node, DataType) {
	assign := p.consume()
	expressionStart := p.peek().Position

	// Collect expression
	expression := p.parseExpressionRoot()

	// Get expression type
	expressionType := p.GetExpressionType(expression)

	// Uncompatible data types
	expressionPosition := dataStructures.CodePos{File: expressionStart.File, Line: expressionStart.Line, StartChar: expressionStart.StartChar, EndChar: p.peekPrevious().Position.EndChar}

	// Print errors
	if expressionType.DType != DT_NoType {
		for i, identifier := range identifierTokens {
			// Variable has a type and it's incompatible with expression
			if variableTypes[i].DType != DT_NoType && !expressionType.Equals(variableTypes[i]) {

				// Assign type to empty list literal
				if variableTypes[i].DType == DT_List && expression.NodeType == NT_List && len(expression.Value.(*ListNode).Nodes) == 0 {
					expression.Value.(*ListNode).DataType.SubType = variableTypes[i].SubType
					continue
				}

				p.newErrorNoMessage(&expressionPosition)
				logger.Error2CodePos(identifierTokens[i].Position, &expressionPosition, fmt.Sprintf("Can't assign expression of type %s to variable %s of type %s.", expressionType, identifier, variableTypes[i]))
			}
		}
	}

	// Operation-Assign nodes
	if assign.TokenType != lexer.TT_KW_Assign {
		nodeType := OperationAssignTokenToNodeType[assign.TokenType]

		for i, identifier := range identifierTokens[:len(identifierTokens)-1] {
			// Transform a += 1 to a = a + 1
			variableNode := &Node{identifierTokens[i].Position, NT_Variable, &VariableNode{identifier.Value, expressionType}}
			newExpression := &Node{assign.Position, NT_Assign, &AssignNode{identifier.Value, &Node{assign.Position, nodeType, &BinaryNode{variableNode, expression, expressionType.DType}}}}

			p.GetExpressionType(newExpression)
			visualize(newExpression, "", true)

			p.appendScope(newExpression)
		}

		// Transform a += 1 to a = a + 1
		variableNode := &Node{identifierTokens[len(identifierTokens)-1].Position, NT_Variable, &VariableNode{identifierTokens[len(identifierTokens)-1].Value, expressionType}}
		newExpression := &Node{assign.Position, nodeType, &BinaryNode{variableNode, expression, expressionType.DType}}

		p.GetExpressionType(newExpression)

		return &Node{assign.Position, NT_Assign, &AssignNode{identifierTokens[len(identifierTokens)-1].Value, newExpression}}, expressionType
	}

	// Assign nodes
	for _, identifier := range identifierTokens[:len(identifierTokens)-1] {
		p.appendScope(&Node{assign.Position, NT_Assign, &AssignNode{identifier.Value, expression}})
	}

	return &Node{assign.Position, NT_Assign, &AssignNode{identifierTokens[len(identifierTokens)-1].Value, expression}}, expressionType
}
