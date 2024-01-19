package parser

import (
	"fmt"
	"math"
	"neko/dataStructures"
	"neko/lexer"
	"strconv"
)

const MINIMAL_PRECEDENCE = -100

func (p *Parser) parseExpression(currentPrecedence int) *Node {
	var left *Node

	// Literal
	if p.peek().TokenType.IsLiteral() {
		var literalValue LiteralValue

		switch p.peek().TokenType {
		case lexer.TT_LT_None:
			literalValue = nil
		case lexer.TT_LT_Bool:
			literalValue = p.peek().Value[0] == '1'
		case lexer.TT_LT_Int:
			literalValue, _ = strconv.ParseInt(p.peek().Value, 10, 64)
		case lexer.TT_LT_Float:
			literalValue, _ = strconv.ParseFloat(p.peek().Value, 64)
		case lexer.TT_LT_String:
			literalValue = p.peek().Value
		}

		left = &Node{p.peek().Position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.consume().TokenType], literalValue}}
		// Sub-Expression
	} else if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		p.consume()
		left = p.parseExpression(MINIMAL_PRECEDENCE)
		p.consume()
		// Unary operators
	} else if p.peek().TokenType.IsUnaryOperator() {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(lexer.TT_OP_Not)) // Unary - has same precedence as !

		// Combine - and int node
		if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).DataType == DT_Int {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(int64)
			left = right
			// Combine - and float node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).DataType == DT_Float {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(float64)
			left = right
			// Combine ! and bool node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Not && right.Value.(*LiteralNode).DataType == DT_Bool {
			right.Value.(*LiteralNode).Value = !right.Value.(*LiteralNode).Value.(bool)
			left = right
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{nil, right, DT_NoType}}
		}

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
		} else if symbol.symbolType == ST_Variable {
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

		if left.NodeType == NT_Literal && right.NodeType == NT_Literal && left.Value.(*LiteralNode).DataType == right.Value.(*LiteralNode).DataType {
			left = combineLiteralNodes(left, right, TokenTypeToNodeType[operator.TokenType], operator.Position)
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{left, right, DT_NoType}}
		}
	}

	return left
}

func operatorPrecedence(operator lexer.TokenType) int {
	switch operator {
	case lexer.TT_OP_And, lexer.TT_OP_Or:
		return 0
	case lexer.TT_OP_Equal, lexer.TT_OP_NotEqual,
		lexer.TT_OP_Lower, lexer.TT_OP_Greater,
		lexer.TT_OP_LowerEqual, lexer.TT_OP_GreaterEqual:
		return 1
	case lexer.TT_OP_Add, lexer.TT_OP_Subtract:
		return 2
	case lexer.TT_OP_Multiply, lexer.TT_OP_Divide:
		return 3
	case lexer.TT_OP_Power, lexer.TT_OP_Modulo:
		return 4
	case lexer.TT_OP_Not:
		return 5
	default:
		panic(fmt.Sprintf("Can't get operator precedence of token type %s.", operator))
	}
}

func (p *Parser) getExpressionType(expression *Node) VariableType {
	if expression.NodeType.IsOperator() {
		// Unary operator
		if expression.Value.(*BinaryNode).Left == nil {
			return p.getExpressionType(expression.Value.(*BinaryNode).Right)
		}

		leftType := p.getExpressionType(expression.Value.(*BinaryNode).Left)
		rightType := p.getExpressionType(expression.Value.(*BinaryNode).Right)

		// Same type on both sides
		if leftType.Equals(rightType) {
			// Logic operators can be used only on booleans
			if expression.NodeType.IsLogicOperator() && (leftType.DataType != DT_Bool || rightType.DataType != DT_Bool) {
				p.newError(expression.Position, fmt.Sprintf("Operator %s can be only used on expressions of type bool.", expression.NodeType))
				return VariableType{DT_Bool, leftType.CanBeNone || rightType.CanBeNone}
			}

			// Only + can be used on strings
			if leftType.DataType == DT_String && expression.NodeType != NT_Add {
				p.newError(expression.Position, fmt.Sprintf("Can't use operator %s on data types %s and %s.", NodeTypeToString[expression.NodeType], leftType, rightType))
				return leftType
			}

			// Comparison operators return boolean
			if expression.NodeType.IsComparisonOperator() {
				return VariableType{DT_Bool, leftType.CanBeNone || rightType.CanBeNone}
			}
			expression.Value.(*BinaryNode).DataType = leftType.DataType
			return leftType
		}

		// Failed to get data type
		if leftType.DataType == DT_NoType || rightType.DataType == DT_NoType {
			return VariableType{DT_NoType, false}
		}

		p.newError(expression.Position, fmt.Sprintf("Operator %s is used on incompatible data types %s and %s.", expression.NodeType, leftType, rightType))
		return VariableType{max(leftType.DataType, rightType.DataType), leftType.CanBeNone || rightType.CanBeNone}
	}

	switch expression.NodeType {
	case NT_Literal:
		return VariableType{expression.Value.(*LiteralNode).DataType, false}
	case NT_Variable:
		return expression.Value.(*VariableNode).VariableType
	case NT_FunctionCall:
		return *expression.Value.(*FunctionCallNode).ReturnType
	}

	panic(fmt.Sprintf("Can't determine expression data type from %s.", NodeTypeToString[expression.NodeType]))
}

func getExpressionPosition(expression *Node, left, right uint) dataStructures.CodePos {
	// Binary node
	if expression.NodeType.IsOperator() {
		binaryNode := expression.Value.(*BinaryNode)

		if binaryNode.Left != nil {
			leftPosition := getExpressionPosition(binaryNode.Left, left, right)
			rightPosition := getExpressionPosition(binaryNode.Right, left, right)

			return dataStructures.CodePos{File: leftPosition.File, Line: leftPosition.Line, StartChar: leftPosition.StartChar, EndChar: rightPosition.EndChar}
		}

		expression = binaryNode.Left
	}

	// Check if node position is outside of bounds of max found position
	position := dataStructures.CodePos{File: expression.Position.File, Line: expression.Position.Line, StartChar: left, EndChar: right}

	if expression.Position.StartChar < left {
		position.StartChar = expression.Position.StartChar
	}

	if expression.Position.EndChar > right {
		position.EndChar = expression.Position.EndChar
	}

	return position
}

func combineLiteralNodes(left, right *Node, parentNodeType NodeType, parentPosition *dataStructures.CodePos) *Node {
	leftLiteral := left.Value.(*LiteralNode)
	rightLiteral := right.Value.(*LiteralNode)

	switch leftLiteral.DataType {
	// Booleans
	case DT_Bool:
		switch parentNodeType {
		case NT_Equal:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, leftLiteral.Value.(bool) == rightLiteral.Value.(bool)}}
		case NT_NotEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, leftLiteral.Value.(bool) != rightLiteral.Value.(bool)}}
		case NT_And:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, leftLiteral.Value.(bool) && rightLiteral.Value.(bool)}}
		case NT_Or:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, leftLiteral.Value.(bool) || rightLiteral.Value.(bool)}}
		}
	// Integers
	case DT_Int:
		var value LiteralValue = nil

		// Arithmetic operations
		switch parentNodeType {
		case NT_Add:
			value = leftLiteral.Value.(int64) + rightLiteral.Value.(int64)
		case NT_Subtract:
			value = leftLiteral.Value.(int64) - rightLiteral.Value.(int64)
		case NT_Multiply:
			value = leftLiteral.Value.(int64) * rightLiteral.Value.(int64)
		case NT_Divide:
			value = leftLiteral.Value.(int64) / rightLiteral.Value.(int64)
		case NT_Power:
			value = powerInt64(leftLiteral.Value.(int64), rightLiteral.Value.(int64))
		case NT_Modulo:
			value = leftLiteral.Value.(int64) % rightLiteral.Value.(int64)
		}

		if value != nil {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, value}}
		}

		// Comparison operators
		switch parentNodeType {
		case NT_Equal:
			value = leftLiteral.Value.(int64) == rightLiteral.Value.(int64)
		case NT_NotEqual:
			value = leftLiteral.Value.(int64) != rightLiteral.Value.(int64)
		case NT_Lower:
			value = leftLiteral.Value.(int64) < rightLiteral.Value.(int64)
		case NT_Greater:
			value = leftLiteral.Value.(int64) > rightLiteral.Value.(int64)
		case NT_LowerEqual:
			value = leftLiteral.Value.(int64) <= rightLiteral.Value.(int64)
		case NT_GreaterEqual:
			value = leftLiteral.Value.(int64) >= rightLiteral.Value.(int64)
		}

		if value != nil {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, value}}
		}

	// Floats
	case DT_Float:
		var value LiteralValue = nil

		// Arithmetic operations
		switch parentNodeType {
		case NT_Add:
			value = leftLiteral.Value.(float64) + rightLiteral.Value.(float64)
		case NT_Subtract:
			value = leftLiteral.Value.(float64) - rightLiteral.Value.(float64)
		case NT_Multiply:
			value = leftLiteral.Value.(float64) * rightLiteral.Value.(float64)
		case NT_Divide:
			value = leftLiteral.Value.(float64) / rightLiteral.Value.(float64)
		case NT_Power:
			value = math.Pow(leftLiteral.Value.(float64), rightLiteral.Value.(float64))
		case NT_Modulo:
			value = math.Mod(leftLiteral.Value.(float64), rightLiteral.Value.(float64))
		}

		if value != nil {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, value}}
		}

		// Comparison operators
		switch parentNodeType {
		case NT_Equal:
			value = leftLiteral.Value.(float64) == rightLiteral.Value.(float64)
		case NT_NotEqual:
			value = leftLiteral.Value.(float64) != rightLiteral.Value.(float64)
		case NT_Lower:
			value = leftLiteral.Value.(float64) < rightLiteral.Value.(float64)
		case NT_Greater:
			value = leftLiteral.Value.(float64) > rightLiteral.Value.(float64)
		case NT_LowerEqual:
			value = leftLiteral.Value.(float64) <= rightLiteral.Value.(float64)
		case NT_GreaterEqual:
			value = leftLiteral.Value.(float64) >= rightLiteral.Value.(float64)
		}

		if value != nil {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, value}}
		}

	// Strings
	case DT_String:
		if parentNodeType == NT_Add {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_String, fmt.Sprintf("%s%s", left.Value.(*LiteralNode).Value, right.Value.(*LiteralNode).Value)}}
		}
	}

	// Invalid operation, can't combine
	return &Node{parentPosition, parentNodeType, &BinaryNode{left, right, DT_NoType}}
}

func powerInt64(base, exponent int64) int64 {
	var result int64 = 1

	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}

	return result
}
