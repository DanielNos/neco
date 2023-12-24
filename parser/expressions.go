package parser

import (
	"fmt"
	"neko/lexer"
)

const MINIMAL_PRECEDENCE = -100

func (p *Parser) parseExpression(currentPrecedence int) *Node {
	var left *Node

	// Literal
	if p.peek().TokenType.IsLiteral() {
		left = &Node{p.peek().Position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.peek().TokenType], p.consume().Value}}
	// Sub-Expression
	} else if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		p.consume()
		left = p.parseExpression(MINIMAL_PRECEDENCE)
		p.consume()
	// Identifiers
	} else if p.peek().TokenType == lexer.TT_Identifier {
		_, exists := p.globalSymbolTable[p.peek().Value]

		if !exists {
			p.newError(p.peek(), fmt.Sprintf("Variable %s is not declared in this scope.", p.peek()))
		}

		left = &Node{p.peek().Position, NT_Variable, &VariableNode{p.consume().Value}}
	// Invalid token
	} else {
		panic(fmt.Sprintf("Invalid token in expression %s.", p.peek()))
	}

	// Operators
	for p.peek().TokenType.IsBinaryOperator() && operatorPrecedence(p.peek().TokenType) >= currentPrecedence {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(operator.TokenType))

		left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{left, right}}
	}

	return left
}

func operatorPrecedence(operator lexer.TokenType) int {
	switch operator {
	case lexer.TT_OP_Equal, lexer.TT_OP_NotEqual,
		 lexer.TT_OP_Lower, lexer.TT_OP_Greater,
		 lexer.TT_OP_LowerEqual, lexer.TT_OP_GreaterEqual:
		return 0
	case lexer.TT_OP_Add, lexer.TT_OP_Subtract:
		return 1
	case lexer.TT_OP_Multiply, lexer.TT_OP_Divide:
		return 2
	case lexer.TT_OP_Power, lexer.TT_OP_Modulo:
		return 3
	default:
		panic(fmt.Sprintf("Can't get operator precedence of token type %s.", operator))
	}
}
