package parser

import (
	"fmt"
	"math"
	data "neco/dataStructures"
	"neco/lexer"
	VM "neco/virtualMachine"
	"strconv"
)

const MINIMAL_PRECEDENCE = -100

func (p *Parser) parseExpressionRoot() *Node {
	expression := p.parseExpression(MINIMAL_PRECEDENCE)

	p.collectConstants(expression)

	return expression
}

func (p *Parser) parseExpression(currentPrecedence int) *Node {
	var left *Node = nil

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Literal
	if p.peek().TokenType.IsLiteral() {
		var literalValue LiteralValue

		switch p.peek().TokenType {
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
		if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).DType == data.DT_Int {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(int64)
			left = right
			// Combine - and float node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).DType == data.DT_Float {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(float64)
			left = right
			// Combine ! and bool node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Not && right.Value.(*LiteralNode).DType == data.DT_Bool {
			right.Value.(*LiteralNode).Value = !right.Value.(*LiteralNode).Value.(bool)
			left = right
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &BinaryNode{nil, right, data.DataType{data.DT_NoType, nil}}}
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
				left = &Node{identifier.Position, NT_Variable, &VariableNode{identifier.Value, data.DataType{data.DT_NoType, nil}}}
			}
			// Function call
		} else if symbol.symbolType == ST_FunctionBucket {
			left = p.parseFunctionCall(symbol, p.consume())
			// Variable
		} else if symbol.symbolType == ST_Variable {
			// Uninitialized variable
			if !symbol.value.(*VariableSymbol).isInitialized {
				p.newError(p.peek().Position, fmt.Sprintf("Variable %s is not initialized.", p.peek()))
			}

			identifierToken := p.consume()

			// List element
			if p.peek().TokenType == lexer.TT_DL_BracketOpen {
				// Consume index
				for p.peek().TokenType == lexer.TT_DL_BracketOpen {
					p.consume()
					if left == nil {
						variable := &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
						left = &Node{identifierToken.Position, NT_ListValue, &BinaryNode{variable, p.parseExpressionRoot(), symbol.value.(*VariableSymbol).VariableType.SubType.(data.DataType)}}
					} else {
						left = &Node{identifierToken.Position, NT_ListValue, &BinaryNode{left, p.parseExpressionRoot(), symbol.value.(*VariableSymbol).VariableType.SubType.(data.DataType)}}
					}

					p.consume()
				}
				// Normal variable
			} else {
				left = &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
			}
		} else {
			left = &Node{p.peek().Position, NT_Variable, &VariableNode{p.consume().Value, data.DataType{data.DT_NoType, nil}}}
		}

		// List
	} else if p.peek().TokenType == lexer.TT_DL_BraceOpen {
		startPosition := p.consume().Position

		// Skip EOC
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}

		// Collect edxpressions
		expressions := []*Node{}
		expressionTypes := map[data.DataType]int{}
		elementType := data.DataType{data.DT_NoType, nil}

		for p.peek().TokenType != lexer.TT_DL_BraceClose {
			// Collect expression
			expressions = append(expressions, p.parseExpressionRoot())

			// Assign list type
			elementType = p.GetExpressionType(expressions[len(expressions)-1])
			expressionTypes[elementType] += 1

			// Consume comma
			if p.peek().TokenType == lexer.TT_DL_Comma {
				p.consume()

				// Skip EOC
				if p.peek().TokenType == lexer.TT_EndOfCommand {
					p.consume()
				}
			} else if p.peek().TokenType == lexer.TT_EndOfCommand {
				p.consume()
			}
		}

		// More than one type in list
		if len(expressionTypes) != 1 {
			// Find type with lowest count
			lowestCount := 999999
			lowestType := data.DataType{}

			for t, count := range expressionTypes {
				if count < lowestCount {
					lowestCount = count
					lowestType = t
				}
			}

			// Find it's expression and print error
			for _, expression := range expressions {
				if p.GetExpressionType(expression).Equals(lowestType) {
					p.newError(expression.Position, "List can't contain elements of multiple data types.")
					break
				}
			}
		}

		left = &Node{startPosition.SetEndPos(p.consume().Position), NT_List, &ListNode{expressions, data.DataType{data.DT_List, elementType}}}
		// Invalid token
	} else {
		panic(fmt.Sprintf("Invalid token in expression %s.", p.peek()))
	}

	// Operators
	for p.peek().TokenType.IsBinaryOperator() && operatorPrecedence(p.peek().TokenType) >= currentPrecedence {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(operator.TokenType))
		nodeType := TokenTypeToNodeType[operator.TokenType]

		// Combine two literals into single node
		if left.NodeType == NT_Literal && right.NodeType == NT_Literal && left.Value.(*LiteralNode).DType == right.Value.(*LiteralNode).DType {
			left = combineLiteralNodes(left, right, nodeType)
			continue
		}

		if right.IsBinaryNode() && operatorNodePrecedence[nodeType] == operatorNodePrecedence[right.NodeType] && nodeType != NT_Power {
			oldLeft := left

			// Rotate nodes
			left = right.Value.(*BinaryNode).Right
			right.Value.(*BinaryNode).Right = right.Value.(*BinaryNode).Left
			right.Value.(*BinaryNode).Left = oldLeft

			left = &Node{left.Position.SetEndPos(right.Position), nodeType, &BinaryNode{right, left, data.DataType{data.DT_NoType, nil}}}
			continue
		}

		left = &Node{left.Position.SetEndPos(right.Position), nodeType, &BinaryNode{left, right, data.DataType{data.DT_NoType, nil}}}
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

func (p *Parser) GetExpressionType(expression *Node) data.DataType {
	if expression.NodeType.IsOperator() {
		// Unary operator
		if expression.Value.(*BinaryNode).Left == nil {
			unaryType := p.GetExpressionType(expression.Value.(*BinaryNode).Right)
			expression.Value.(*BinaryNode).DataType = unaryType
			return unaryType
		}

		leftType := p.GetExpressionType(expression.Value.(*BinaryNode).Left)
		rightType := p.GetExpressionType(expression.Value.(*BinaryNode).Right)

		// Error in one of types
		if leftType.DType == data.DT_NoType || rightType.DType == data.DT_NoType {
			return data.DataType{data.DT_NoType, nil}
		}

		// Same type on both sides
		if leftType.Equals(rightType) {
			// Logic operators can be used only on booleans
			if expression.NodeType.IsLogicOperator() && (leftType.DType != data.DT_Bool || rightType.DType != data.DT_Bool) {
				p.newError(expression.Position, fmt.Sprintf("Operator %s can be only used on expressions of type bool.", expression.NodeType))
				return data.DataType{data.DT_Bool, nil}
			}

			// Comparison operators return boolean
			if expression.NodeType.IsComparisonOperator() {
				expression.Value.(*BinaryNode).DataType = data.DataType{data.DT_Bool, nil}
				return data.DataType{data.DT_Bool, nil}
			}

			// Only + can be used on strings and lists
			if (leftType.DType == data.DT_String || leftType.DType == data.DT_List) && expression.NodeType != NT_Add {
				p.newError(expression.Position, fmt.Sprintf("Can't use operator %s on data types %s and %s.", NodeTypeToString[expression.NodeType], leftType, rightType))
				return data.DataType{data.DT_NoType, nil}
			}

			// Return left type
			if leftType.DType != data.DT_NoType {
				expression.Value.(*BinaryNode).DataType = leftType
				return leftType
			}

			// Return right type
			if rightType.DType != data.DT_NoType {
				expression.Value.(*BinaryNode).DataType = rightType
				return rightType
			}

			// Neither have type
			return leftType
		}

		// Failed to get data type
		if leftType.DType == data.DT_NoType || rightType.DType == data.DT_NoType {
			return data.DataType{data.DT_NoType, nil}
		}

		p.newError(expression.Position, fmt.Sprintf("Operator %s is used on incompatible data types %s and %s.", expression.NodeType, leftType, rightType))

		return data.DataType{data.DT_NoType, nil}
	}

	switch expression.NodeType {
	case NT_Literal:
		return data.DataType{expression.Value.(*LiteralNode).DType, nil}
	case NT_Variable:
		return expression.Value.(*VariableNode).DataType
	case NT_FunctionCall:
		return *expression.Value.(*FunctionCallNode).ReturnType
	case NT_List:
		return expression.Value.(*ListNode).DataType
	case NT_ListValue:
		return p.GetExpressionType(expression.Value.(*BinaryNode).Left)
	}

	panic(fmt.Sprintf("Can't determine expression data type from %s.", NodeTypeToString[expression.NodeType]))
}

func (p *Parser) collectConstants(expression *Node) {
	// Check operator children
	if expression.IsBinaryNode() {
		p.collectConstants(expression.Value.(*BinaryNode).Left)
		p.collectConstants(expression.Value.(*BinaryNode).Right)
		// Collect literal
	} else if expression.NodeType == NT_Literal {
		literalNode := expression.Value.(*LiteralNode)

		switch literalNode.DType {
		case data.DT_Int:
			p.IntConstants[literalNode.Value.(int64)] = -1
		case data.DT_Float:
			p.FloatConstants[literalNode.Value.(float64)] = -1
		case data.DT_String:
			p.StringConstants[literalNode.Value.(string)] = -1
		}
	}
}

func combineLiteralNodes(left, right *Node, parentNodeType NodeType) *Node {
	leftLiteral := left.Value.(*LiteralNode)
	rightLiteral := right.Value.(*LiteralNode)

	switch leftLiteral.DType {
	// Booleans
	case data.DT_Bool:
		switch parentNodeType {
		case NT_Equal:
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) == rightLiteral.Value.(bool)}}
		case NT_NotEqual:
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) != rightLiteral.Value.(bool)}}
		case NT_And:
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) && rightLiteral.Value.(bool)}}
		case NT_Or:
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) || rightLiteral.Value.(bool)}}
		}
	// Integers
	case data.DT_Int:
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
			value = VM.PowerInt64(leftLiteral.Value.(int64), rightLiteral.Value.(int64))
		case NT_Modulo:
			value = leftLiteral.Value.(int64) % rightLiteral.Value.(int64)
		}

		if value != nil {
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Int, value}}
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
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, value}}
		}

	// Floats
	case data.DT_Float:
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
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Float, value}}
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
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, value}}
		}

	// Strings
	case data.DT_String:
		if parentNodeType == NT_Add {
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_String, fmt.Sprintf("%s%s", left.Value.(*LiteralNode).Value, right.Value.(*LiteralNode).Value)}}
		}
	}

	// Invalid operation, can't combine
	return &Node{left.Position.SetEndPos(right.Position), parentNodeType, &BinaryNode{left, right, data.DataType{data.DT_NoType, nil}}}
}
