package parser

import (
	"fmt"
	data "neco/dataStructures"
	"neco/lexer"
	"neco/logger"
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

		if nodeType == NT_Ternary {
			// Right side has to have two expressions
			if right.NodeType != NT_TernaryBranches {
				p.newError(GetExpressionPosition(right), "Right side of the ternary operator ?? need to have two expressions separated by a \":\".")
			} else {
				p.collectConstant(right.Value.(*TypedBinaryNode).Right)
				p.collectConstant(right.Value.(*TypedBinaryNode).Left)
			}

			left = p.createBinaryNode(operator.Position, nodeType, left, right)
			continue
		}

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
	// Unwrap option
	if expression.NodeType == NT_Unwrap {
		unwrappedNode := expression.Value.(*Node)
		unwrappedNodeType := GetExpressionType(unwrappedNode)

		// Expression can't be unwrapped
		if unwrappedNodeType.Type != data.DT_Option {
			p.newError(GetExpressionPosition(unwrappedNode), "Can't unwrap an expression with type "+unwrappedNodeType.String()+".")
		}

		return unwrappedNodeType.SubType.(*data.DataType)
	}

	// Match statement
	if expression.NodeType == NT_Match {
		return expression.Value.(*MatchNode).DataType
	}

	// Operators
	if expression.NodeType.IsOperator() {
		return p.deriveOperatorType(expression)
	}

	return GetExpressionType(expression)
}

func (p *Parser) deriveOperatorType(expression *Node) *data.DataType {
	binaryNode := expression.Value.(*TypedBinaryNode)

	// Node has it's type stored already
	if binaryNode.DataType.Type != data.DT_Unknown {
		return binaryNode.DataType
	}

	// Unary operator
	if binaryNode.Left == nil {
		return GetExpressionType(binaryNode.Right)
	}

	// Ternary expression
	if expression.NodeType == NT_Ternary {
		leftType := p.deriveType(binaryNode.Left)

		// Left side isn't bool
		if leftType.Type != data.DT_Bool {
			p.newError(GetExpressionPosition(binaryNode.Left), "Left side of ternary operator ?? has to be of type bool.")
		}

		ternaryBranches := binaryNode.Right.Value.(*TypedBinaryNode)
		leftBranchType := p.deriveType(ternaryBranches.Left)
		rightBranchType := p.deriveType(ternaryBranches.Right)

		// Branches have different types
		if !leftBranchType.Equals(rightBranchType) {
			p.newError(GetExpressionPosition(binaryNode.Right), "Both branches of ternary operator ?? have to be of the same type. Left is "+leftBranchType.String()+", right is "+rightBranchType.String()+".")
		}

		ternaryBranches.DataType = leftBranchType
		binaryNode.DataType = leftBranchType
		return leftBranchType
	}

	// Type of tranches of ternary operator can't be derived
	if expression.NodeType == NT_TernaryBranches {
		p.newError(GetExpressionPosition(expression), "Unexpected expression. Are you missing an \"??\" operator?")
		return &data.DataType{data.DT_Unknown, nil}
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

	if expression.NodeType == NT_UnpackOrDefault {
		// Right side of ?! can't be none
		if rightType.Type == data.DT_None {
			p.newError(GetExpressionPosition(binaryNode.Right), "Expression on the right side of ?! operator can't be none.")
		} else if rightType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(binaryNode.Right), "Expression on the right side of ?! operator can't be possibly none.")
		}

		// Check if left and right type is compatible
		if leftType.Type == data.DT_Option {
			if !leftType.CanBeAssigned(rightType) {
				p.newError(GetExpressionPosition(binaryNode.Left), "Both sides of operator ?! have to have the same type. Left is "+leftType.String()+", right is "+rightType.String()+".")
			}
			// Left has to be option or none
		} else if leftType.Type != data.DT_None {
			p.newError(GetExpressionPosition(binaryNode.Left), "Left side of operator ?! has to be an option type.")
		}

		binaryNode.DataType = rightType
		return rightType
	}

	// Compatible data types, check if operator is allowed
	if leftType.CanBeAssigned(rightType) {
		// Can't do any operations on options without unwrapping
		if leftType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left), "Options need to be unwrapped or matched to access their values.")
		}
		if rightType.Type == data.DT_Option {
			p.newError(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right), "Options need to be unwrapped or matched to access their values.")
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

func operatorPrecedence(operator lexer.TokenType) int {
	switch operator {
	case lexer.TT_OP_Ternary, lexer.TT_DL_Colon:
		return 0
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

		return &Node{identifierToken.Position.SetEndPos(p.peek().Position), NT_Enum, &EnumNode{identifierToken.Value, symbol.value.(map[string]int64)[p.consume().Value]}}
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
			node := &Node{identifierToken.Position.SetEndPos(p.consume().Position), NT_ObjectField, &ObjectFieldNode{left, property.number, property.dataType}}

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

func GetExpressionType(expression *Node) *data.DataType {
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
		return GetExpressionType(expression.Value.(*TypedBinaryNode).Left).SubType.(*data.DataType)
	case NT_Enum:
		return &data.DataType{data.DT_Enum, expression.Value.(*EnumNode).Identifier}
	case NT_Object:
		return &data.DataType{data.DT_Object, expression.Value.(*ObjectNode).Identifier}
	case NT_ObjectField:
		return expression.Value.(*ObjectFieldNode).DataType
	case NT_Set:
		return expression.Value.(*ListNode).DataType
	case NT_Unwrap:
		return GetExpressionType(expression.Value.(*Node)).SubType.(*data.DataType)
	case NT_IsNone:
		return &data.DataType{data.DT_Bool, nil}
	case NT_Match:
		return expression.Value.(*MatchNode).DataType
	case NT_Ternary:
		return GetExpressionType(expression.Value.(*TypedBinaryNode).Right)
	case NT_TernaryBranches:
		return expression.Value.(*TypedBinaryNode).DataType
	}

	panic("Can't determine expression data type from " + NodeTypeToString[expression.NodeType] + fmt.Sprintf(" (%d)", expression.NodeType) + ".")
}

func GetExpressionPosition(expression *Node) *data.CodePos {
	if expression.NodeType.IsOperator() {
		return GetExpressionPosition(expression.Value.(*TypedBinaryNode).Left).SetEndPos(GetExpressionPosition(expression.Value.(*TypedBinaryNode).Right))
	}

	if expression.NodeType == NT_Unwrap || expression.NodeType == NT_IsNone {
		return GetExpressionPosition(expression.Value.(*Node)).SetEndPos(expression.Position)
	}

	return expression.Position
}
