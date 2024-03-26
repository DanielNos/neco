package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
	"neco/logger"
)

func (p *Parser) parseVariableDeclaration(constant bool) *Node {
	startPosition := p.peek().Position

	// Collect data type
	variableType := p.parseType()

	// Collect identifiers
	variableNodes, variableIdentifiers := p.parseVariableIdentifiers(variableType)

	// Create node
	declareNode := &Node{startPosition.SetEndPos(variableNodes[len(variableNodes)-1].Position), NT_VariableDeclare, &VariableDeclareNode{variableType, constant, variableIdentifiers}}

	// End
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		// var has to be assigned to
		if variableType.Type == data.DT_NoType {
			startPosition.EndChar = p.peekPrevious().Position.EndChar
			p.newError(startPosition, "Variables declared using keyword var have to have an expression assigned to them, so a data type can be derived from it.")
		}
		p.consume()
		// Assign
	} else if p.peek().TokenType == lexer.TT_KW_Assign {
		// Push declare node
		node := declareNode
		p.appendScope(node)

		// Parse expression and collect type
		var expressionType data.DataType
		declareNode, expressionType = p.parseAssign(variableNodes, startPosition)

		// Change variable type if no was provided
		if variableType.Type == data.DT_NoType {
			variableType = expressionType
			node.Value.(*VariableDeclareNode).DataType = expressionType
		}
	}

	// Insert symbols
	for _, id := range variableIdentifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{variableType, declareNode.NodeType == NT_Assign, constant}})
	}

	return declareNode
}

func (p *Parser) parseAssign(assignedStatements []*Node, startOfStatement *data.CodePos) (*Node, data.DataType) {
	assign := p.consume()
	expressionStart := p.peek().Position

	// Collect expression
	expression := p.parseExpressionRoot()
	expressionType := p.GetExpressionType(expression)

	// Uncompatible data types
	expressionPosition := data.CodePos{expressionStart.File, expressionStart.StartLine, expressionStart.EndLine, expressionStart.StartChar, p.peekPrevious().Position.EndChar}

	// Print errors
	if expressionType.Type != data.DT_NoType {
		for _, assignedTo := range assignedStatements {
			variableType := p.GetExpressionType(assignedTo)

			// Leaf type of expression is set
			if expressionType.SubType != nil {
				// Check if variable can be assigned expression
				if !variableType.CanBeAssigned(expressionType) {
					// Variable doesn't have type yet (declared using var)
					if variableType.Type == data.DT_NoType {
						variableType = expressionType
						// Invalid type
					} else {
						p.newErrorNoMessage()
						logger.Error2CodePos(assignedTo.Position, &expressionPosition, "Can't assign expression of type "+expressionType.String()+" to variable of type "+variableType.String()+".")
					}
				}
				// Leaf type of expression is not set => use variable type
			} else {
				// Assignin list<?> or set<?> expression to a var variable, so type can't be determined
				if variableType.Type == data.DT_NoType {
					if expressionType.Type == data.DT_Set {
						p.newErrorNoMessage()
						logger.Error2CodePos(assignedTo.Position, &expressionPosition, "Can't assign expression of type set<?> to variable declared using keyword var. Replace var with required type.")
					} else if expressionType.Type == data.DT_List {
						p.newErrorNoMessage()
						logger.Error2CodePos(assignedTo.Position, &expressionPosition, "Can't assign expression of type list<?> to variable declared using keyword var. Replace var with required type.")
					}
					// Set expressions sub-type to variable sub-type (both list and set use *ListNode for elements)
				} else {
					expression.Value.(*ListNode).DataType = variableType
				}
			}
		}
	}

	// Operation-assign nodes need to be edited
	if assign.TokenType != lexer.TT_KW_Assign {
		nodeType := OperationAssignTokenToNodeType[assign.TokenType]

		// Transform assigned expressions in the following way: a += 1 to a = a + 1
		for _, assignedStatement := range assignedStatements[:len(assignedStatements)-1] {
			generatedNode := &Node{assign.Position, nodeType, &TypedBinaryNode{assignedStatement, expression, expressionType}}
			p.appendScope(&Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{[]*Node{assignedStatement}, generatedNode}})
		}

		generatedNode := &Node{assign.Position, nodeType, &TypedBinaryNode{assignedStatements[len(assignedStatements)-1], expression, expressionType}}
		return &Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{[]*Node{assignedStatements[len(assignedStatements)-1]}, generatedNode}}, expressionType
	}

	return &Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{assignedStatements, expression}}, expressionType
}
