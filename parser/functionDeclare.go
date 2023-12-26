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
	scope := &ScopeNode{p.scopeCounter, []*Node{}}
	p.scopeNodeStack.Push(scope)
	p.scopeCounter++
	p.symbolTableStack.Push(map[string]*Symbol{})

	p.consume()

	// Collect parameters
	if p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		p.parseParameters()
	}

	p.consume()

	// Parse body
	p.parseScope(false)

	// Leave scope
	p.scopeNodeStack.Pop()
	p.symbolTableStack.Pop()

	return &Node{start, NT_FunctionDeclare, &FunctionDeclareNode{identifier}}
}

func (p *Parser) parseParameters() {
	for p.peek().TokenType != lexer.TT_EndOfFile {
		dataType := TokenTypeToDataType[p.consume().TokenType]
		identifier := p.consume().Value

		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{dataType, false, false}})

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
	}
}
