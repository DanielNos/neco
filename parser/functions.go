package parser

import (
	"fmt"
	"neko/lexer"
)

func (p *Parser) parseFunctionDeclare() *Node {
	start := p.consume().Position
	
	// Check for redeclaration
	symbol := p.findSymbol(p.peek().Value)

	if symbol != nil {
		p.newError(p.peek(), fmt.Sprintf("Redeclaration of symbol %s.", p.peek().Value))
	}

	// Collect name
	identifier := p.consume().Value

	// Enter scope
	p.enterScope()

	p.consume()

	// Collect parameters
	parameters := p.parseParameters()

	p.consume()

	// Collect return type
	var returnType *DataType = nil

	if p.peek().TokenType == lexer.TT_KW_returns {
		p.consume()
		returnTypeToken := TokenTypeToDataType[p.consume().TokenType]
		returnType = &returnTypeToken
	}

	// Parse body
	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false)}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.symbolTableStack.Pop()


	// Insert function symbol
	p.insertSymbol(identifier, &Symbol{ST_Function, &FunctionSymbol{parameters}})

	return &Node{start, NT_FunctionDeclare, &FunctionDeclareNode{identifier, parameters, returnType, body}}
}

func (p *Parser) parseParameters() []Parameter {
	var paremeters = []Parameter{}

	for p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		// Collect data typa and identifier
		dataType := TokenTypeToDataType[p.consume().TokenType]
		identifier := p.consume().Value

		// Create parameter and symbol
		paremeters = append(paremeters, Parameter{dataType, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{dataType, false, false}})
	}

	return paremeters
}

func (p *Parser) parseFunctionCall(functionSymbol *Symbol) *Node {
	identifier := p.consume()

	// Symbol not provided
	if functionSymbol == nil {
		functionSymbol = p.getGlobalSymbol(identifier.Value)
	}

	// Undeclared function
	if functionSymbol == nil {
		p.newError(identifier, fmt.Sprintf("Use of undeclared function %s.", identifier.Value))
	}

	// Collect arguments
	arguments := p.parseArguments()

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{identifier.Value, arguments}}
}

func (p *Parser) parseArguments() []*Node {
	p.consume()
	var arguments = []*Node{}

	for p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		arguments = append(arguments, p.parseExpression(MINIMAL_PRECEDENCE))
	}
	p.consume()

	return arguments
}
