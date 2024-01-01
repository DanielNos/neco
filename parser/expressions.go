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
		left = &Node{p.peek().Position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.peek().TokenType], p.consume().Value}}
	// Sub-Expression
	} else if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		p.consume()
		left = p.parseExpression(MINIMAL_PRECEDENCE)
		p.consume()
	// Unary operators
	} else if p.peek().TokenType.IsUnaryOperator() {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(operator.TokenType))

		// Combine - and int/float nodes
		if right.nodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && (right.value.(*LiteralNode).dataType == DT_Int || right.value.(*LiteralNode).dataType == DT_Float){
			right.value.(*LiteralNode).value = fmt.Sprintf("-%s", right.value.(*LiteralNode).value)
			left = right
		// Combine ! and bool nodes
		} else if  right.nodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Not && right.value.(*LiteralNode).dataType == DT_Bool {
			if right.value.(*LiteralNode).value[0] == '0' {
				right.value.(*LiteralNode).value = "1"
			} else {
				right.value.(*LiteralNode).value = "0"
			}
			
			left = right
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{nil, right}}
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

		if left.nodeType == NT_Literal && right.nodeType == NT_Literal && left.value.(*LiteralNode).dataType == right.value.(*LiteralNode).dataType{
			left = combineLiteralNodes(left, right, TokenTypeToNodeType[operator.TokenType], operator.Position)
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{left, right}}
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
	if expression.nodeType.IsOperator() {
		// Unary operator
		if expression.value.(*BinaryNode).left == nil {
			return p.getExpressionType(expression.value.(*BinaryNode).right)
		}

		leftType := p.getExpressionType(expression.value.(*BinaryNode).left)
		rightType := p.getExpressionType(expression.value.(*BinaryNode).right)

		// Same type on both sides
		if leftType.Equals(rightType) {
			// Logic operators can be used only on booleans
			if expression.nodeType.IsLogicOperator() && (leftType.dataType != DT_Bool || rightType.dataType != DT_Bool) {
				p.newError(expression.position, fmt.Sprintf("Operator %s can be only used on expressions of type bool.", expression.nodeType))
				return VariableType{DT_Bool, leftType.canBeNone || rightType.canBeNone}
			}

			// Only + can be used on strings
			if leftType.dataType == DT_String && expression.nodeType != NT_Add {
				p.newError(expression.position, fmt.Sprintf("Can't use operator %s on data types %s and %s.", NodeTypeToString[expression.nodeType], leftType, rightType))
				return leftType
			}

			// Comparison operators return boolean
			if expression.nodeType.IsComparisonOperator() {
				return VariableType{DT_Bool, leftType.canBeNone || rightType.canBeNone}
			}

			return leftType
		}

		// Failed to get data type
		if leftType.dataType == DT_NoType || rightType.dataType == DT_NoType {
			return VariableType{DT_NoType, false}
		}

		p.newError(expression.position, fmt.Sprintf("Operator %s is used on incompatible data types %s and %s.", expression.nodeType, leftType, rightType))
		return VariableType{max(leftType.dataType, rightType.dataType), leftType.canBeNone || rightType.canBeNone}
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

func getExpressionPosition(expression *Node, left, right uint) dataStructures.CodePos {
	// Binary node
	if expression.nodeType.IsOperator() {
		binaryNode := expression.value.(*BinaryNode)

		if binaryNode.left != nil {
			leftPosition := getExpressionPosition(binaryNode.left, left, right)
			rightPosition := getExpressionPosition(binaryNode.right, left, right)
	
			return dataStructures.CodePos{File: leftPosition.File, Line: leftPosition.Line, StartChar: leftPosition.StartChar, EndChar: rightPosition.EndChar}
		}

		expression = binaryNode.left
	}

	// Check if node position is outside of bounds of max found position
	position := dataStructures.CodePos{File: expression.position.File, Line: expression.position.Line, StartChar: left, EndChar: right}

	if expression.position.StartChar < left {
		position.StartChar = expression.position.StartChar
	}

	if expression.position.EndChar > right {
		position.EndChar = expression.position.EndChar
	}

	return position
}

func combineLiteralNodes(left, right *Node, parentNodeType NodeType, parentPosition *dataStructures.CodePos) *Node {
	leftLiteral := left.value.(*LiteralNode)
	rightLiteral := right.value.(*LiteralNode)

	switch leftLiteral.dataType {
	// Booleans
	case DT_Bool:
		leftValue := leftLiteral.value[0] == '1'
		rightValue := rightLiteral.value[0] == '1'

		switch parentNodeType {
		case NT_Equal:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue == rightValue)}}
		case NT_NotEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue != rightValue)}}
		case NT_And:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue && rightValue)}}
		case NT_Or:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue || rightValue)}}
		}
	// Integers
	case DT_Int:
		leftValue, _ := strconv.ParseInt(leftLiteral.value, 10, 64)
		rightValue, _ := strconv.ParseInt(rightLiteral.value, 10, 64)

		switch parentNodeType {
		case NT_Add:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", leftValue + rightValue)}}
		case NT_Subtract:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", leftValue - rightValue)}}
		case NT_Multiply:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", leftValue * rightValue)}}
		case NT_Divide:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", leftValue / rightValue)}}
		case NT_Power:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", powerInt64(leftValue, rightValue))}}
		case NT_Modulo:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Int, fmt.Sprintf("%d", leftValue % rightValue)}}
		case NT_Equal:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue == rightValue)}}
		case NT_NotEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue != rightValue)}}
		case NT_Lower:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue < rightValue)}}
		case NT_Greater:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue > rightValue)}}
		case NT_LowerEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue <= rightValue)}}
		case NT_GreaterEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue >= rightValue)}}
		}
	// Floats
	case DT_Float:
		leftValue, _ := strconv.ParseFloat(leftLiteral.value, 64)
		rightValue, _ := strconv.ParseFloat(rightLiteral.value, 64)

		switch parentNodeType {
		case NT_Add:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", leftValue + rightValue)}}
		case NT_Subtract:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", leftValue - rightValue)}}
		case NT_Multiply:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", leftValue * rightValue)}}
		case NT_Divide:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", leftValue / rightValue)}}
		case NT_Power:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", math.Pow(leftValue, rightValue))}}
		case NT_Modulo:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Float, fmt.Sprintf("%.75g", math.Mod(leftValue, rightValue))}}
		case NT_Equal:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue == rightValue)}}
		case NT_NotEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue != rightValue)}}
		case NT_Lower:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue < rightValue)}}
		case NT_Greater:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue > rightValue)}}
		case NT_LowerEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue <= rightValue)}}
		case NT_GreaterEqual:
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_Bool, boolToString(leftValue >= rightValue)}}
		}
	// Strings
	case DT_String:
		if parentNodeType == NT_Add {
			return &Node{parentPosition, NT_Literal, &LiteralNode{DT_String, fmt.Sprintf("%s%s", left.value, right.value)}}
		}
	}

	// Invalid operation, can't combine
	return &Node{parentPosition, parentNodeType, &BinaryNode{left, right}}
}

func powerInt64(base, exponent int64) int64 {
	var result int64 = 1

	for exponent > 0 {
		if exponent % 2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}

	return result
}

func boolToString(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
