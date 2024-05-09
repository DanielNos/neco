package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

const MINIMAL_PRECEDENCE = -100

type dataTypeCount struct {
	DataType *data.DataType
	Count    int
}

func (p *Parser) parseExpressionRoot() *Node {
	expression := p.parseExpression(MINIMAL_PRECEDENCE)
	VisualizeNode(expression)
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
		left = p.parseLiteral()

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
		left = p.parseIdentifier(true)

		// List
	} else if p.peek().TokenType == lexer.TT_DL_BracketOpen {
		left = p.parseEnumeration(false)

		// Set
	} else if p.peek().TokenType == lexer.TT_DL_BraceOpen {
		left = p.parseEnumeration(true)

		// List/Set with a type specified
	} else if p.peek().TokenType.IsCompositeType() {
		specifiedType := p.parseType()

		// Expression after type isn't a list/set
		if p.peek().TokenType != lexer.TT_DL_BraceOpen && p.peek().TokenType != lexer.TT_DL_BracketOpen {
			left = p.parseExpression(currentPrecedence)
			p.newError(GetExpressionPosition(left), "Expected expression of the type "+specifiedType.String()+".")
			// Try to set the type of the expression
		} else {
			left = p.parseEnumeration(p.peek().TokenType == lexer.TT_DL_BraceOpen)
			expressionType := left.Value.(*ListNode).DataType.Copy()
			expressionType.TryCompleteFrom(specifiedType)

			// Type of expression after type hint is incompatible with it
			if !specifiedType.CanBeAssigned(expressionType) {
				p.newError(GetExpressionPosition(left), "Expression after type hint "+specifiedType.String()+" has thew wrong type "+left.Value.(*ListNode).DataType.String()+".")
			}
			left.Value.(*ListNode).DataType = specifiedType
		}

		// Match statement
	} else if p.peek().TokenType == lexer.TT_KW_match {
		left = p.parseMatchExpression()

		// Ternary operator
	} else if p.peek().TokenType == lexer.TT_OP_Ternary {
		right := p.parseExpressionRoot()

		// Right side has to have two expressions
		if right.NodeType != NT_TernaryBranches {
			p.newError(GetExpressionPosition(right), "Right side of the ternary operator ?? need to have two expressions separated by a \":\".")
		} else {
			p.collectConstant(right.Value.(*TypedBinaryNode).Right)
			p.collectConstant(right.Value.(*TypedBinaryNode).Left)
		}

		return p.createBinaryNode(p.consume().Position, NT_Ternary, left, right)
	
		// Invalid token
	} else {
		panic("Invalid token in expression " + p.peek().String() + ".")
	}

	// Left is unwrapped
	if p.peek().TokenType == lexer.TT_OP_Not {
		left = &Node{p.consume().Position, NT_Unwrap, left}
		// Left is checked for none
	} else if p.peek().TokenType == lexer.TT_OP_QuestionMark {
		left = &Node{p.consume().Position, NT_IsNone, left}
	}

	// Operators
	for p.peek().TokenType.IsBinaryOperator() && operatorPrecedence(p.peek().TokenType) >= currentPrecedence {
		operator := p.consume()

		// Parse right side of expression
		right := p.parseExpression(operatorPrecedence(operator.TokenType))
		nodeType := TokenTypeToNodeType[operator.TokenType]

		// Combine two literals into single node
		if p.optimize && left.NodeType == NT_Literal && right.NodeType == NT_Literal && left.Value.(*LiteralNode).PrimitiveType == right.Value.(*LiteralNode).PrimitiveType {
			left = combineLiteralNodes(left, right, nodeType)
			continue
		}

		// Right node is binary node with same precedence => rotate nodes so they are left-to-right associated (except power, which is right-to-left associated)
		if right.IsBinaryNode() && operatorNodePrecedence[nodeType] == operatorNodePrecedence[right.NodeType] && nodeType != NT_Power {
			left = p.rotateNodes(left, right, operator.Position, nodeType)
			continue
		}

		left = p.createBinaryNode(operator.Position, nodeType, left, right)
	}

	return left
}

func (p *Parser) rotateNodes(left, right *Node, position *data.CodePos, nodeType NodeType) *Node {
	oldLeft := left
	p.collectConstant(left)

	// Rotate nodes
	left = right.Value.(*TypedBinaryNode).Right
	right.Value.(*TypedBinaryNode).Right = right.Value.(*TypedBinaryNode).Left
	right.Value.(*TypedBinaryNode).Left = oldLeft

	// Create node
	left = p.createBinaryNode(position, nodeType, right, left)

	// Swap node types and positions
	left.NodeType, right.NodeType = right.NodeType, left.NodeType
	left.Position, right.Position = right.Position, left.Position

	return left
}

func (p *Parser) createBinaryNode(position *data.CodePos, nodeType NodeType, left, right *Node) *Node {
	// Store constants
	p.collectConstant(left)
	p.collectConstant(right)
	return &Node{position, nodeType, &TypedBinaryNode{left, right, &data.DataType{data.DT_Unknown, nil}}}
}

func operatorPrecedence(operator lexer.TokenType) int {
	switch operator {
	case lexer.TT_OP_UnpackOrDefault:
		return 1
	case lexer.TT_OP_Or:
		return 2
	case lexer.TT_OP_And:
		return 3
	case lexer.TT_OP_Equal, lexer.TT_OP_NotEqual,
		lexer.TT_OP_Lower, lexer.TT_OP_Greater,
		lexer.TT_OP_LowerEqual, lexer.TT_OP_GreaterEqual,
		lexer.TT_OP_In:
		return 4
	case lexer.TT_OP_Add, lexer.TT_OP_Subtract:
		return 5
	case lexer.TT_OP_Multiply, lexer.TT_OP_Divide:
		return 6
	case lexer.TT_OP_Power, lexer.TT_OP_Modulo:
		return 7
	case lexer.TT_OP_Not:
		return 8
	case lexer.TT_OP_Dot:
		return 9
	default:
		panic("Can't get operator precedence of token type " + operator.String() + ".")
	}
}

func (p *Parser) parseIdentifier(isInExpression bool) *Node {
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
		if !isInExpression {
			symbol.value.(*VariableSymbol).isInitialized = true
		}
		return p.parseVariable(symbol)
		// Enum
	} else if symbol.symbolType == ST_Enum {
		identifierToken := p.consume()
		p.consume() // .

		return &Node{identifierToken.Position.Combine(p.peek().Position), NT_Enum, &EnumNode{identifierToken.Value, symbol.value.(map[string]int64)[p.consume().Value]}}
		// Struct
	} else if symbol.symbolType == ST_Struct {
		return p.parseStructLiteral(symbol.value.(map[string]PropertySymbol))
	}

	return nil
}

func (p *Parser) parseVariable(symbol *Symbol) *Node {
	variableSymbol := symbol.value.(*VariableSymbol)

	// Uninitialized variable
	if !variableSymbol.isInitialized {
		p.newError(p.peek().Position, "Variable "+p.peek().String()+" is not initialized.")
	}

	identifierToken := p.consume()

	// List element
	if p.peek().TokenType == lexer.TT_DL_BracketOpen {
		// Consume index expression
		for p.peek().TokenType == lexer.TT_DL_BracketOpen {
			p.consume() // [
			variable := &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
			listValue := &Node{identifierToken.Position, NT_ListValue, &TypedBinaryNode{variable, p.parseExpressionRoot(), symbol.value.(*VariableSymbol).VariableType.SubType.(*data.DataType)}}
			p.consume() // ]

			return listValue
		}
		// Object field
	} else if p.peek().TokenType == lexer.TT_OP_Dot {
		variableNode := &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, variableSymbol.VariableType}}
		return p.parseObjectField(variableNode, variableSymbol.VariableType, identifierToken)
	}

	// Normal variable
	return &Node{identifierToken.Position, NT_Variable, &VariableNode{identifierToken.Value, symbol.value.(*VariableSymbol).VariableType}}
}

func (p *Parser) parseObjectField(left *Node, leftType *data.DataType, identifierToken *lexer.Token) *Node {
	p.consume() // .

	// Can access properties of structs only
	if leftType.Type != data.DT_Object {
		p.newError(p.peek().Position, "Can't access a property of "+identifierToken.String()+", because it's not a struct.")
	} else {
		// Find struct definition
		structName := leftType.SubType.(string)
		structSymbol := p.getGlobalSymbol(structName)

		// Check if field exists
		property, propertyExists := structSymbol.value.(map[string]PropertySymbol)[p.peek().Value]

		if !propertyExists {
			p.newError(p.peek().Position, "Struct "+structName+" doesn't have a property "+p.consume().Value+".")
		} else {
			node := &Node{identifierToken.Position.Combine(p.consume().Position), NT_ObjectField, &ObjectFieldNode{left, property.number, property.dataType}}

			// Another field access
			if p.peek().TokenType == lexer.TT_OP_Dot {
				return p.parseObjectField(node, property.dataType, p.peekPrevious())
			}

			return node
		}
	}

	return left
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

func (p *Parser) parseEnumeration(isSet bool) *Node {
	// Select collect properties for the structure
	structureName := "List"
	nodeType := NT_List
	dataType := data.DT_List
	closingToken := lexer.TT_DL_BracketClose

	if isSet {
		structureName = "Set"
		nodeType = NT_Set
		dataType = data.DT_Set
		closingToken = lexer.TT_DL_BraceClose
	}

	startPosition := p.consume().Position

	// Skip EOC
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Collect elements
	expressions := []*Node{}
	expressionTypes := map[string]*dataTypeCount{}
	elementType := &data.DataType{data.DT_Unknown, nil}

	for p.peek().TokenType != closingToken {
		// Collect expression
		expressions = append(expressions, p.parseExpressionRoot())

		// Assign list type
		elementType = GetExpressionType(expressions[len(expressions)-1])
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

	return &Node{startPosition.Combine(p.consume().Position), nodeType, &ListNode{expressions, &data.DataType{dataType, elementType}}}
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
			if GetExpressionType(expression).CanBeAssigned(lowestType) {
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

func GetExpressionPosition(expression *Node) *data.CodePos {
	if expression.NodeType.IsOperator() {
		return GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left).Combine(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right))
	}

	if expression.NodeType == NT_Unwrap || expression.NodeType == NT_IsNone {
		return GetExpressionPosition(expression.Value.(*Node)).Combine(expression.Position)
	}

	return expression.Position
}
