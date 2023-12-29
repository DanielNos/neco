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
		symbol := p.findSymbol(p.peek().Value)

		// Undeclared symbol
		if symbol == nil {
			identifier := p.consume()
			
			// Undeclared function
			if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
				p.newError(identifier.Position, fmt.Sprintf("Function %s is not declared in this scope.", identifier.Value))
				left = p.parseFunctionCall(nil, identifier)
			// Undeclared variable
			} else {
				p.newError(identifier.Position, fmt.Sprintf("Variable %s is not declared in this scope.", identifier.Value))
				left = &Node{identifier.Position, NT_Variable, &VariableNode{identifier.Value, VariableType{DT_NoType, false}}}
			}
		// Function call
		} else if symbol.symbolType == ST_Function {
			left = p.parseFunctionCall(symbol, p.consume())
		// Variable
		} else if symbol.symbolType == ST_Variable{
			// Uninitialized variable
			if !symbol.value.(*VariableSymbol).isInitialized {
				p.newError(p.peek().Position, fmt.Sprintf("Variable %s is not initialized.", p.peek()))
			}
			left = &Node{p.peek().Position, NT_Variable, &VariableNode{p.consume().Value, symbol.value.(*VariableSymbol).variableType}}
		} else {
			left = &Node{p.peek().Position, NT_Variable, &VariableNode{p.consume().Value, VariableType{DT_NoType, false}}}
		}

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

func (p *Parser) getExpressionType(expression *Node) VariableType {
	if expression.nodeType.IsOperator() {
		leftType := p.getExpressionType(expression.value.(*BinaryNode).left)
		rightType := p.getExpressionType(expression.value.(*BinaryNode).right)

		// Same type on both sides
		if leftType.Equals(rightType) {
			if leftType.dataType == DT_String && expression.nodeType > NT_Add && expression.nodeType <= NT_Modulo {
				p.newError(expression.position, fmt.Sprintf("Can't use operator %s on data types %s and %s.", NodeTypeToString[expression.nodeType], leftType, rightType))
			}

			return leftType
		}

		// Failed to get data type
		if leftType.dataType == DT_NoType || rightType.dataType == DT_NoType {
			return VariableType{DT_NoType, false}
		}

	}

	switch expression.nodeType {
	case NT_Literal:
		return VariableType{expression.value.(*LiteralNode).dataType, false}
	case NT_Variable:
		return expression.value.(*VariableNode).variableType
	case NT_FunctionCall:
		return *expression.value.(*FunctionCallNode).returnType
	}

	panic(fmt.Sprintf("Can't determine expression data type from %s.", NodeTypeToString[expression.nodeType]))
}
