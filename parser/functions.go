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
		p.newError(p.peek().Position, fmt.Sprintf("Redeclaration of symbol %s.", p.peek().Value))
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
	returnType := VariableType{DT_NoType, false}

	if p.peek().TokenType == lexer.TT_KW_returns {
		p.consume()
		returnType.dataType = TokenTypeToDataType[p.consume().TokenType]
	}

	// Parse body
	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false)}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.symbolTableStack.Pop()

	// Insert function symbol
	p.insertSymbol(identifier, &Symbol{ST_Function, &FunctionSymbol{parameters, returnType}})

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
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{VariableType{dataType, false}, false}})
	}

	return paremeters
}

func (p *Parser) parseFunctionCall(functionSymbol *Symbol, identifier *lexer.Token) *Node {
	// Collect arguments
	arguments := p.parseArguments(functionSymbol.value.(*FunctionSymbol).parameters)

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{identifier.Value, arguments, &functionSymbol.value.(*FunctionSymbol).returnType}}
}

func (p *Parser) parseArguments(paramters []Parameter) []*Node {
	p.consume()
	var arguments = []*Node{}

	// Collect arguments
	for p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		arguments = append(arguments, p.parseExpression(MINIMAL_PRECEDENCE))
	}
	p.consume()

	return arguments
}
