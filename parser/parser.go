package parser

import (
	"fmt"
	data "neco/dataStructures"
	"neco/errors"
	"neco/lexer"
	"neco/logger"
	"os"
	"strings"
)

var TokenTypeToDataType = map[lexer.TokenType]data.PrimitiveType{
	lexer.TT_KW_bool: data.DT_Bool,
	lexer.TT_LT_Bool: data.DT_Bool,

	lexer.TT_KW_int: data.DT_Int,
	lexer.TT_LT_Int: data.DT_Int,

	lexer.TT_KW_flt:   data.DT_Float,
	lexer.TT_LT_Float: data.DT_Float,

	lexer.TT_KW_str:    data.DT_String,
	lexer.TT_LT_String: data.DT_String,

	lexer.TT_KW_list: data.DT_List,
	lexer.TT_KW_set:  data.DT_Set,
}

type Parser struct {
	tokens []*lexer.Token

	tokenIndex int

	scopeCounter   int
	scopeNodeStack *data.Stack

	stack_symbolTableStack *data.Stack

	functions     []*FunctionSymbol
	functionIndex int

	ErrorCount      uint
	totalErrorCount uint

	IntConstants    map[int64]int
	FloatConstants  map[float64]int
	StringConstants map[string]int
}

func NewParser(tokens []*lexer.Token, previousErrors uint) Parser {
	return Parser{
		tokens: tokens,

		tokenIndex: 0,

		scopeCounter:   0,
		scopeNodeStack: data.NewStack(),

		stack_symbolTableStack: data.NewStack(),

		functions:     []*FunctionSymbol{},
		functionIndex: 0,

		ErrorCount:      0,
		totalErrorCount: previousErrors,

		IntConstants:    map[int64]int{},
		FloatConstants:  map[float64]int{},
		StringConstants: map[string]int{},
	}
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

func (p *Parser) newError(position *data.CodePos, message string) {
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

func (p *Parser) newErrorNoMessage() {
	if p.ErrorCount+p.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	p.ErrorCount++

	// Too many errors
	if p.ErrorCount+p.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Semantic analysis has aborted due to too many errors. It has failed with %d errors.", p.ErrorCount))
	}
}

func (p *Parser) enterScope() {
	p.stack_symbolTableStack.Push(symbolTable{})
	p.scopeNodeStack.Push(&ScopeNode{p.scopeCounter, []*Node{}})
	p.scopeCounter++
}

func (p *Parser) leaveScope() {
	p.scopeNodeStack.Pop()
	p.stack_symbolTableStack.Pop()
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

	// Collect enums, structs and function headers
	p.collectGlobals()

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
	for identifier, symbol := range p.stack_symbolTableStack.Bottom.Value.(symbolTable) {
		// Try to find function bucket symbol
		if symbol.symbolType == ST_FunctionBucket {
			// Check if every function in the bucket was ever called
			for _, functionSymbol := range symbol.value.(symbolTable) {
				if !functionSymbol.value.(*FunctionSymbol).everCalled {
					logger.Warning("Function " + identifier + " was never called.")
				}
			}
		}

	}

	return module
}

func (p *Parser) parseScope(enterScope, packInNode bool) interface{} {
	// Consume opening brace
	opening := p.peek().Position

	if p.peek().TokenType == lexer.TT_DL_BraceOpen {
		p.consume()
	}

	// Enter or use current scope
	if enterScope {
		p.enterScope()
	}
	scope := p.scopeNodeStack.Top.Value.(*ScopeNode)

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
	case lexer.TT_KW_var, lexer.TT_KW_bool, lexer.TT_KW_int, lexer.TT_KW_flt, lexer.TT_KW_str, lexer.TT_KW_list, lexer.TT_KW_set:
		return p.parseVariableDeclaration(false)

	case lexer.TT_KW_const:
		p.consume()
		return p.parseVariableDeclaration(true)

	// Function declaration
	case lexer.TT_KW_fun:
		return p.parseFunctionDeclaration()

	// Leave scope
	case lexer.TT_DL_BraceClose:
		// Pop scope
		if p.scopeNodeStack.Size > 1 {
			if enteredScope {
				p.leaveScope()
			}
			p.consume()
			// Root scope
		} else {
			p.newError(p.consume().Position, "Unexpected closing brace in root scope.")
		}

		return nil

	// Identifier
	case lexer.TT_Identifier:
		return p.parseIdentifierStatement()

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

	// ForEach
	case lexer.TT_KW_forEach:
		return p.parseForEach()

	// Scope
	case lexer.TT_DL_BraceOpen:
		return p.parseScope(true, true).(*Node)

	// Break
	case lexer.TT_KW_break:
		return &Node{p.consume().Position, NT_Break, nil}

	// Struct, enum
	case lexer.TT_KW_struct, lexer.TT_KW_enum:
		// Skip over enums and structs, because they were registered in collectGlobals()
		for p.peek().TokenType != lexer.TT_DL_BraceClose {
			p.consume()
		}
		p.consume()

		return p.parseStatement(enteredScope)

	// Delete
	case lexer.TT_KW_delete:
		return p.parseDelete()

	// Skip EOCs
	case lexer.TT_EndOfCommand:
		p.consume()
		return p.parseStatement(enteredScope)

	// Return no node for EndOfFile token
	case lexer.TT_EndOfFile:
		return nil
	}

	panic(p.peek().Position.String() + " Unexpected token " + p.peek().TokenType.String() + " \"" + p.consume().String() + "\".")
}

func (p *Parser) parseType() *data.DataType {
	// Convert current token to data type
	variableType := &data.DataType{TokenTypeToDataType[p.peek().TokenType], nil}

	// Token is not a data type keyword => it's enum or struct
	if variableType.Type == data.DT_Unknown {
		symbol := p.getGlobalSymbol(p.peek().Value)

		// Neither primitive or user defined type => type can't be determined
		if symbol == nil {
			p.consume()
			return variableType // DT_Unknown
		}

		// Symbol is a struct
		if symbol.symbolType == ST_Struct {
			variableType.Type = data.DT_Object
			// Symbol is a enum
		} else if symbol.symbolType == ST_Enum {
			variableType.Type = data.DT_Enum
		}

		// Set sub-type to struct/enum name
		variableType.SubType = p.peek().Value
	}

	// Create data type
	p.consume()

	// Insert subtype to list data type
	if variableType.Type == data.DT_List || variableType.Type == data.DT_Set {
		p.consume()
		variableType.SubType = p.parseType()
		p.consume()
	}

	return variableType
}

func (p *Parser) consumeEOCs() {
	for p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}
}

func (p *Parser) parseDelete() *Node {
	position := p.consume().Position

	// Missing expression
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.newError(position, "Expected deleted expression after keyword delete.")
		return &Node{position, NT_Delete, nil}
	}

	// Collect expression
	expression := p.parseExpressionRoot()

	// Check if expression can be deleted
	switch expression.NodeType {
	// Remove variable from scope
	case NT_Variable:
		// Look up variable
		_, exists := p.stack_symbolTableStack.Top.Value.(symbolTable)[expression.Value.(*VariableNode).Identifier]

		if exists {
			delete(p.stack_symbolTableStack.Top.Value.(symbolTable), expression.Value.(*VariableNode).Identifier)
		}

	// Accept also In and ListValue
	case NT_In, NT_ListValue:

	// Invalid target
	default:
		p.newError(GetExpressionPosition(expression), "Expression can't be deleted.")
	}

	return &Node{position, NT_Delete, expression}
}
