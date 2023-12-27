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

	var parameters = []Parameter{}

	// Collect parameters
	if p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		parameters = p.parseParameters()
	}

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

	return &Node{start, NT_FunctionDeclare, &FunctionDeclareNode{identifier, parameters, returnType, body}}
}

func (p *Parser) parseParameters() []Parameter {
	var paremeters = []Parameter{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		dataType := TokenTypeToDataType[p.consume().TokenType]
		identifier := p.consume().Value

		paremeters = append(paremeters, Parameter{dataType, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{dataType, false, false}})

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
	}

	return paremeters
}
