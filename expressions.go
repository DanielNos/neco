package main

import "fmt"

const MINIMAL_PRECEDENCE = -100

func (p *Parser) parseExpression(currentPrecedence int) *Node {
	var left *Node

	// Literal
	if p.peek().tokenType.IsLiteral() {
		left = &Node{p.peek().position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.peek().tokenType], p.consume().value}}
	// Sub-Expression
	} else if p.peek().tokenType == TT_DL_ParenthesisOpen {
		p.consume()
		left = p.parseExpression(MINIMAL_PRECEDENCE)
		p.consume()
	// Identifiers
	} else if p.peek().tokenType == TT_Identifier {
		_, exists := p.globalSymbolTable[p.peek().value]

		if !exists {
			p.newError(p.peek(), fmt.Sprintf("Variable %s is not declared in this scope.", p.peek()))
		}

		left = &Node{p.peek().position, NT_Variable, &VariableNode{p.consume().value}}
	// Invalid token
	} else {
		panic(fmt.Sprintf("Invalid token in expression %s.", p.peek()))
	}

	// Operators
	for p.peek().tokenType.IsBinaryOperator() && operatorPrecedence(p.peek().tokenType) >= currentPrecedence {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(operator.tokenType))

		left = &Node{operator.position, TokenTypeToNodeType[operator.tokenType], &BinaryNode{left, right}}
	}

	return left
}

func operatorPrecedence(operator TokenType) int {
	switch operator {
	case TT_OP_Equal, TT_OP_NotEqual, TT_OP_Lower, TT_OP_Greater, TT_OP_LowerEqual, TT_OP_GreaterEqual:
		return 0
	case TT_OP_Add, TT_OP_Subtract:
		return 1
	case TT_OP_Multiply, TT_OP_Divide:
		return 2
	case TT_OP_Power, TT_OP_Modulo:
		return 3
	default:
		panic(fmt.Sprintf("Can't get operator precedence of token type %s.", operator))
	}
}
