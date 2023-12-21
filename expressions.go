package main

func (p *Parser) parseExpression() *Node {
	var left *Node

	if p.peek().tokenType.IsLiteral() {
		left = &Node{p.peek().position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.peek().tokenType], p.consume().value}}
	}
	
	return left
}

func operatorPrecedence(operator TokenType) int {
	switch operator {
	case TT_OP_Add, TT_OP_Subtract:
		return 0
	case TT_OP_Multiply, TT_OP_Divide:
		return 1
	case TT_OP_Power, TT_OP_Modulo:
		return 2
	}

	return -1000
}
