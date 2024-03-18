package parser

import (
	"fmt"
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
		if variableType.DType == data.DT_NoType {
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
		if variableType.DType == data.DT_NoType {
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

	// Get expression type
	expressionType := p.GetExpressionType(expression)

	// Uncompatible data types
	expressionPosition := data.CodePos{expressionStart.File, expressionStart.StartLine, expressionStart.EndLine, expressionStart.StartChar, p.peekPrevious().Position.EndChar}

	// Print errors
	if expressionType.DType != data.DT_NoType {
		for _, assignedTo := range assignedStatements {
			variableType := p.GetExpressionType(assignedTo)

			// Find leaf data type
			leafType := expressionType

			for leafType.DType == data.DT_List {
				leafType = leafType.SubType.(data.DataType)
			}

			// Check if variable can be assigned expression
			if leafType.DType != data.DT_NoType && !variableType.Equals(expressionType) {
				p.newErrorNoMessage()
				logger.Error2CodePos(assignedTo.Position, &expressionPosition, fmt.Sprintf("Can't assign expression of type %s to variable of type %s.", expressionType, variableType))
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
