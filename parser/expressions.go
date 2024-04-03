package parser

import (
	"fmt"
	"math"
	data "neco/dataStructures"
	"neco/lexer"
	"neco/logger"
	VM "neco/virtualMachine"
	"strconv"
)

const MINIMAL_PRECEDENCE = -100

type dataTypeCount struct {
	DataType *data.DataType
	Count    int
}

func (p *Parser) parseExpressionRoot() *Node {
	expression := p.parseExpression(MINIMAL_PRECEDENCE)
	p.collectConstant(expression)
	p.deriveType(expression)

	return expression
}

func (p *Parser) parseExpression(currentPrecedence int) *Node {
	var left *Node = nil

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Literal
	if p.peek().TokenType.IsLiteral() {
		// Parse literal value from token value string
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
		case lexer.TT_LT_None:
			literalValue = ""
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
		if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).PrimitiveType == data.DT_Int {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(int64)
			left = right
			// Combine - and float node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Subtract && right.Value.(*LiteralNode).PrimitiveType == data.DT_Float {
			right.Value.(*LiteralNode).Value = -right.Value.(*LiteralNode).Value.(float64)
			left = right
			// Combine ! and bool node
		} else if right.NodeType == NT_Literal && operator.TokenType == lexer.TT_OP_Not && right.Value.(*LiteralNode).PrimitiveType == data.DT_Bool {
			right.Value.(*LiteralNode).Value = !right.Value.(*LiteralNode).Value.(bool)
			left = right
		} else {
			left = &Node{operator.Position, TokenTypeToNodeType[operator.TokenType], &TypedBinaryNode{nil, right, &data.DataType{data.DT_Unknown, nil}}}
		}

		// Identifiers
	} else if p.peek().TokenType == lexer.TT_Identifier {
		left = p.parseIdentifier()

		// List
	} else if p.peek().TokenType == lexer.TT_DL_BracketOpen {
		left = p.parseEnumeration("List", NT_List, data.DT_List)

		// Set
	} else if p.peek().TokenType == lexer.TT_DL_BraceOpen {
		left = p.parseEnumeration("Set", NT_Set, data.DT_Set)

		// Invalid token
	} else {
		panic("Invalid token in expression " + p.peek().String() + ".")
	}

	// Operators
	for p.peek().TokenType.IsBinaryOperator() && operatorPrecedence(p.peek().TokenType) >= currentPrecedence {
		operator := p.consume()
		right := p.parseExpression(operatorPrecedence(operator.TokenType))
		nodeType := TokenTypeToNodeType[operator.TokenType]

		// Combine two literals into single node
		if p.optimize && left.NodeType == NT_Literal && right.NodeType == NT_Literal && left.Value.(*LiteralNode).PrimitiveType == right.Value.(*LiteralNode).PrimitiveType {
			left = combineLiteralNodes(left, right, nodeType)
			continue
		}

		// Right node is binary node with same precedence => rotate nodes so they are left-to-right associated (except power, which is right-to-left associated)
		if right.IsBinaryNode() && operatorNodePrecedence[nodeType] == operatorNodePrecedence[right.NodeType] && nodeType != NT_Power {
			oldLeft := left
			p.collectConstant(left)

			// Rotate nodes
			left = right.Value.(*TypedBinaryNode).Right
			right.Value.(*TypedBinaryNode).Right = right.Value.(*TypedBinaryNode).Left
			right.Value.(*TypedBinaryNode).Left = oldLeft

			// Create node
			left = p.createBinaryNode(operator.Position, nodeType, right, left)

			// Swap node types and positions
			left.NodeType, right.NodeType = right.NodeType, left.NodeType
			left.Position, right.Position = right.Position, left.Position

			continue
		}

		left = p.createBinaryNode(operator.Position, nodeType, left, right)
	}

	return left
}

func (p *Parser) createBinaryNode(position *data.CodePos, nodeType NodeType, left, right *Node) *Node {
	// Store constants
	p.collectConstant(left)
	p.collectConstant(right)
	return &Node{position, nodeType, &TypedBinaryNode{left, right, &data.DataType{data.DT_Unknown, nil}}}
}

func (p *Parser) deriveType(expression *Node) *data.DataType {
	// Operators
	if expression.NodeType.IsOperator() {
		binaryNode := expression.Value.(*TypedBinaryNode)

		// Node has it's type stored already
		if binaryNode.DataType.Type != data.DT_Unknown {
			return binaryNode.DataType
		}

		// Unary operator
		if binaryNode.Left == nil {
			return p.GetExpressionType(binaryNode.Right)
		}

		// Collect left and right node data types
		leftType := p.deriveType(binaryNode.Left)
		rightType := p.deriveType(binaryNode.Right)

		// Error in one of types
		if leftType.Type == data.DT_Unknown || rightType.Type == data.DT_Unknown {
			return &data.DataType{data.DT_Unknown, nil}
		}

		// In operator has to be used on set with correct type
		if expression.NodeType == NT_In {
			// Right type isn't a set
			if rightType.Type != data.DT_Set {
				p.newError(GetExpressionPosition(binaryNode.Right), "Right side of operator \"in\" has to be a set.")
				// Left type isn't set's sub-type
			} else if !rightType.SubType.(*data.DataType).CanBeAssigned(leftType) {
				p.newErrorNoMessage()
				logger.Error2CodePos(GetExpressionPosition(binaryNode.Left), GetExpressionPosition(binaryNode.Right), "Left expression type ("+leftType.String()+") doesn't match the set element type ("+rightType.SubType.(*data.DataType).String()+").")
			}
			binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
			return binaryNode.DataType
		}

		// Compatible data types, check if operator is allowed
		if leftType.CanBeAssigned(rightType) {
			// Can't do any operations on options without unpacking
			if leftType.Type == data.DT_Option {
				p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left), "Options need to be unpacked to access their values. Use unpack() or match.")
			}
			if rightType.Type == data.DT_Option {
				p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right), "Options need to be unpacked to access their values. Use unpack() or match.")
			}

			// Logic operators can be used only on booleans
			if expression.NodeType.IsLogicOperator() && (leftType.Type != data.DT_Bool || rightType.Type != data.DT_Bool) {
				p.newError(expression.Position, "Operator "+expression.NodeType.String()+" can be only used on expressions of type bool.")
				binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
				return binaryNode.DataType
			}

			// Comparison operators return boolean
			if expression.NodeType.IsComparisonOperator() {
				binaryNode.DataType = &data.DataType{data.DT_Bool, nil}
				return binaryNode.DataType
			}

			// Can't do non-comparison operations on enums
			if leftType.Type == data.DT_Enum || rightType.Type == data.DT_Enum {
				p.newError(expression.Position, "Operator "+expression.NodeType.String()+" can't be used on enum constants.")
				return &data.DataType{data.DT_Unknown, nil}
			}

			// Only + can be used on strings and lists
			if (leftType.Type == data.DT_String || leftType.Type == data.DT_List) && expression.NodeType != NT_Add {
				p.newError(expression.Position, "Can't use operator "+expression.NodeType.String()+" on data types "+leftType.String()+" and "+rightType.String()+".")
				return &data.DataType{data.DT_Unknown, nil}
			}

			// Return left type
			if leftType.Type != data.DT_Unknown {
				binaryNode.DataType = leftType
				return leftType
			}

			// Return right type
			if rightType.Type != data.DT_Unknown {
				binaryNode.DataType = rightType
				return rightType
			}

			// Neither have type
			return leftType
		}

		// Left or right doesn't have a type
		if leftType.Type == data.DT_Unknown || rightType.Type == data.DT_Unknown {
			return leftType
		}

		// Failed to determine data type
		p.newError(expression.Position, "Operator "+expression.NodeType.String()+" is used on incompatible data types "+leftType.String()+" and "+rightType.String()+".")
		return &data.DataType{data.DT_Unknown, nil}
	}

	return p.GetExpressionType(expression)
}

func operatorPrecedence(operator lexer.TokenType) int {
	switch operator {
	case lexer.TT_OP_And, lexer.TT_OP_Or:
		return 0
	case lexer.TT_OP_Equal, lexer.TT_OP_NotEqual,
		lexer.TT_OP_Lower, lexer.TT_OP_Greater,
		lexer.TT_OP_LowerEqual, lexer.TT_OP_GreaterEqual,
		lexer.TT_OP_In:
		return 1
	case lexer.TT_OP_Add, lexer.TT_OP_Subtract:
		return 2
	case lexer.TT_OP_Multiply, lexer.TT_OP_Divide:
		return 3
	case lexer.TT_OP_Power, lexer.TT_OP_Modulo:
		return 4
	case lexer.TT_OP_Not:
		return 5
	case lexer.TT_OP_Dot:
		return 6
	default:
		panic("Can't get operator precedence of token type " + operator.String() + ".")
	}
}

func (p *Parser) parseIdentifier() *Node {
	symbol := p.findSymbol(p.peek().Value)

	// Undeclared symbol
	if symbol == nil {
		identifier := p.consume()

		// Undeclared function
		if p.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
			p.newError(identifier.Position, "Function "+identifier.Value+" is not declared in this scope.")
			return p.parseFunctionCall(nil, identifier)
			// Undeclared struct
		} else if p.peek().TokenType == lexer.TT_DL_BraceOpen {
			p.newError(identifier.Position, "Struct "+identifier.Value+" is not defined in this scope.")

			p.consume() // {
			p.parseAnyProperties()
			p.consume() // }

			return &Node{identifier.Position, NT_Object, &ObjectNode{identifier.Value, []*Node{}}}
			// Undeclared variable
		} else {
			p.newError(identifier.Position, "Variable "+identifier.Value+" is not declared in this scope.")
			return &Node{identifier.Position, NT_Variable, &VariableNode{identifier.Value, &data.DataType{data.DT_Unknown, nil}}}
		}
		// Function call
	} else if symbol.symbolType == ST_FunctionBucket {
		return p.parseFunctionCall(symbol, p.consume())
		// Variable
	} else if symbol.symbolType == ST_Variable {
		// Uninitialized variable
		if !symbol.value.(*VariableSymbol).isInitialized {
			p.newError(p.peek().Position, "Variable "+p.peek().String()+" is not initialized.")
		}

		identifierToken := p.consume()

		// List element
		if p.peek().TokenType == lexer.TT_DL_BracketOpen {
			// Consume index
			for p.peek().TokenType == lexer.TT_DL_BracketOpen {
				p.consume() // [
				variable := &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
				listValue := &Node{identifierToken.Position, NT_ListValue, &TypedBinaryNode{variable, p.parseExpressionRoot(), symbol.value.(*VariableSymbol).VariableType.SubType.(*data.DataType)}}
				p.consume() // ]

				return listValue
			}
			// Struct property
		} else if p.peek().TokenType == lexer.TT_OP_Dot {
			p.consume()
			// Can access properties of structs only
			if symbol.value.(*VariableSymbol).VariableType.Type != data.DT_Object {
				p.newError(p.peek().Position, "Can't access a property of "+identifierToken.String()+", because it's not a struct.")
			} else {
				// Find struct definition
				structName := symbol.value.(*VariableSymbol).VariableType.SubType.(string)
				structSymbol := p.getGlobalSymbol(structName)

				// Check if field exists
				property, propertyExists := structSymbol.value.(map[string]PropertySymbol)[p.peek().Value]

				if !propertyExists {
					p.newError(p.peek().Position, "Struct "+structName+" doesn't have a property "+p.consume().Value+".")
				} else {
					return &Node{identifierToken.Position.SetEndPos(p.consume().Position), NT_StructField, &ObjectFieldNode{identifierToken.Value, property.number, property.dataType}}
				}
			}

			// Normal variable
		} else {
			return &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
		}
		// Enum
	} else if symbol.symbolType == ST_Enum {
		identifierToken := p.consume()
		p.consume() // .

		return &Node{identifierToken.Position.SetEndPos(p.peek().Position), NT_Enum, &EnumNode{identifierToken.Value, symbol.value.(map[string]int64)[p.consume().Value]}}
		// Struct
	} else if symbol.symbolType == ST_Struct {
		return p.parseStructLiteral(symbol.value.(map[string]PropertySymbol))
	}

	return nil
}

func (p *Parser) parseStructLiteral(properties map[string]PropertySymbol) *Node {
	identifier := p.consume()
	p.StringConstants[identifier.Value] = -1

	p.consume() // {
	p.consumeEOCs()

	var propertyValues []*Node

	// Collect named properties
	if p.peek().TokenType == lexer.TT_Identifier && p.peekNext().TokenType == lexer.TT_DL_Colon {
		propertyValues = p.parseKeyedProperties(properties, identifier.Value)
	} else {
		propertyValues = p.parseProperties(properties, identifier)
	}

	p.consume() // }

	return &Node{identifier.Position.SetEndPos(p.peekPrevious().Position), NT_Object, &ObjectNode{identifier.Value, propertyValues}}
}

func (p *Parser) parseKeyedProperties(properties map[string]PropertySymbol, structName string) []*Node {
	propertyValues := map[string]*Node{}

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Field doesn't have a key
		if p.peek().TokenType != lexer.TT_Identifier || p.peekNext().TokenType != lexer.TT_DL_Colon {

			p.newError(p.peek().Position, "All values have to have keys in keyed struct creation.")

			// Collect expression
			p.parseExpressionRoot()

		} else {
			// Collect key
			propertyName := p.consume()
			p.consume()

			// Look up property
			property, exists := properties[propertyName.Value]

			// It doesn't exist
			if !exists {
				p.newError(propertyName.Position, "Struct "+structName+" doesn't have a field "+propertyName.Value+".")
				// It exists
			} else {
				// Check if property is already assigned
				_, isReassigned := propertyValues[propertyName.Value]

				// It's already assigned
				if isReassigned {
					p.newError(propertyName.Position, "Field "+propertyName.Value+" is already assigned.")
				}
			}

			// Collect expression
			expression := p.parseExpressionRoot()

			// Check if expression has correct type
			if exists {
				expressionType := p.GetExpressionType(expression)
				if !property.dataType.CanBeAssigned(expressionType) {
					p.newError(expression.Position, "Field "+propertyName.Value+" of struct "+structName+" has type "+property.dataType.String()+", but is assigned expression of type "+expressionType.String()+".")
				}
			}

			// Store field value
			propertyValues[propertyName.Value] = expression
		}

		// More fields
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()

			// Collect EOCs
			p.consumeEOCs()
		} else {
			for p.peek().TokenType != lexer.TT_DL_BraceClose {
				p.newError(p.peek().Position, "Unexpected token after struct field value.")
			}
		}
	}

	// Change order of values to match property order
	orderedValues := make([]*Node, len(properties))

	for key, property := range properties {
		propertyValue, exists := propertyValues[key]

		if exists {
			orderedValues[property.number] = propertyValue
		}
	}

	return orderedValues
}

func (p *Parser) parseProperties(properties map[string]PropertySymbol, structName *lexer.Token) []*Node {
	// Make properties linear
	orderedProperties := make([]PropertySymbol, len(properties))
	orderedPropertyNames := make([]string, len(properties))

	for key, property := range properties {
		orderedPropertyNames[property.number] = key
		orderedProperties[property.number] = property
	}

	// Collect field values
	propertyValues := make([]*Node, len(properties))
	propertyIndex := 0

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Too many fields
		if propertyIndex == len(properties) {
			p.newError(p.peek().Position, "Struct "+structName.Value+fmt.Sprintf(" has %d fields, but %d values were provided.", len(properties), propertyIndex+1))
			p.parseExpressionRoot()
			// Collect field value
		} else {
			// Collect field expression
			expression := p.parseExpressionRoot()
			expressionType := p.GetExpressionType(expression)

			// Check type
			if !orderedProperties[propertyIndex].dataType.CanBeAssigned(expressionType) {
				p.newError(expression.Position, "Property "+orderedPropertyNames[propertyIndex]+" of struct "+structName.Value+" has type "+orderedProperties[propertyIndex].dataType.String()+", but was assigned expression of type "+expressionType.String()+".")
			}

			// Store property
			propertyValues[propertyIndex] = expression
			propertyIndex++
		}

		// Consume comma
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
		}

		// Consume EOCs
		p.consumeEOCs()
	}

	if propertyIndex < len(properties) {
		p.newError(structName.Position, "Struct "+structName.Value+fmt.Sprintf(" has %d fields, but only %d fields were assigned.", len(properties), propertyIndex))
	}

	return propertyValues
}

func (p *Parser) parseAnyProperties() {
	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Collect property expression
		p.parseExpressionRoot()

		// Consume comma
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
		}

		// Consume EOCs
		p.consumeEOCs()
	}
}

func (p *Parser) parseEnumeration(structureName string, nodeType NodeType, dataType data.PrimitiveType) *Node {
	startPosition := p.consume().Position

	// Skip EOC
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Collect elements
	expressions := []*Node{}
	expressionTypes := map[string]*dataTypeCount{}
	elementType := &data.DataType{data.DT_Unknown, nil}

	for p.peek().TokenType != lexer.TT_DL_BracketClose {
		// Collect expression
		expressions = append(expressions, p.parseExpressionRoot())

		// Assign list type
		elementType = p.GetExpressionType(expressions[len(expressions)-1])
		signature := elementType.Signature()

		typeAndCount, exists := expressionTypes[signature]
		if !exists {
			expressionTypes[signature] = &dataTypeCount{elementType, 1}
		} else {
			typeAndCount.Count++
		}

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

	// Check if all list elements have the same type and set element type to the most common type
	if len(expressionTypes) > 1 {
		elementType = p.checkElementTypes(expressionTypes, expressions, structureName)
	}

	return &Node{startPosition.SetEndPos(p.consume().Position), nodeType, &ListNode{expressions, &data.DataType{dataType, elementType}}}
}

func (p *Parser) checkElementTypes(expressionTypes map[string]*dataTypeCount, expressions []*Node, structureName string) *data.DataType {
	// Find type with lowest count and highest count
	lowestCount := 999999
	lowestType := &data.DataType{}

	highestCount := 0
	highestType := &data.DataType{}

	for _, typeCount := range expressionTypes {
		// Update lowest count
		if typeCount.Count < lowestCount {
			lowestCount = typeCount.Count
			lowestType = typeCount.DataType
		}
		// Update highest count
		if typeCount.Count > highestCount {
			highestCount = typeCount.Count
			highestType = typeCount.DataType
		}
	}

	// There are more than one data types
	if len(expressionTypes) > 1 {
		// Find it's expression and print error
		for _, expression := range expressions {
			if p.GetExpressionType(expression).CanBeAssigned(lowestType) {
				p.newError(expression.Position, structureName+" can't contain elements of multiple data types.")
				break
			}
		}
	}

	return highestType
}

func (p *Parser) collectConstant(node *Node) {
	// Collect literal
	if node.NodeType == NT_Literal {
		literalNode := node.Value.(*LiteralNode)

		switch literalNode.PrimitiveType {
		case data.DT_Int:
			p.IntConstants[literalNode.Value.(int64)] = -1
		case data.DT_Float:
			p.FloatConstants[literalNode.Value.(float64)] = -1
		case data.DT_String:
			p.StringConstants[literalNode.Value.(string)] = -1
		}
		// Collect enum value
	} else if node.NodeType == NT_Enum {
		p.IntConstants[node.Value.(*EnumNode).Value] = -1
	}
}

func combineLiteralNodes(left, right *Node, parentNodeType NodeType) *Node {
	leftLiteral := left.Value.(*LiteralNode)
	rightLiteral := right.Value.(*LiteralNode)

	switch leftLiteral.PrimitiveType {
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
			return &Node{left.Position.SetEndPos(right.Position), NT_Literal, &LiteralNode{data.DT_String, left.Value.(*LiteralNode).Value.(string) + right.Value.(*LiteralNode).Value.(string)}}
		}
	}

	// Invalid operation, can't combine
	return &Node{left.Position.SetEndPos(right.Position), parentNodeType, &TypedBinaryNode{left, right, &data.DataType{data.DT_Unknown, nil}}}
}

func (p *Parser) GetExpressionType(expression *Node) *data.DataType {
	// Binary nodes store their type
	if expression.NodeType.IsOperator() {
		return expression.Value.(*TypedBinaryNode).DataType
	}

	switch expression.NodeType {
	case NT_Literal:
		return &data.DataType{expression.Value.(*LiteralNode).PrimitiveType, nil}
	case NT_Variable:
		return expression.Value.(*VariableNode).DataType
	case NT_FunctionCall:
		return expression.Value.(*FunctionCallNode).ReturnType
	case NT_List:
		return expression.Value.(*ListNode).DataType
	case NT_ListValue:
		return p.GetExpressionType(expression.Value.(*TypedBinaryNode).Left).SubType.(*data.DataType)
	case NT_Enum:
		return &data.DataType{data.DT_Enum, expression.Value.(*EnumNode).Identifier}
	case NT_Object:
		return &data.DataType{data.DT_Object, expression.Value.(*ObjectNode).Identifier}
	case NT_StructField:
		return expression.Value.(*ObjectFieldNode).DataType
	case NT_Set:
		return expression.Value.(*ListNode).DataType
	}

	panic("Can't determine expression data type from " + NodeTypeToString[expression.NodeType] + ".")
}

func GetExpressionPosition(expression *Node) *data.CodePos {
	if expression.NodeType.IsOperator() {
		return GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left).SetEndPos(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right))
	}

	return expression.Position
}
