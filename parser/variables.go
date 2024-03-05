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
	identifierTokens := []*lexer.Token{}
	identifiers := []string{}
	variableTypes := []data.DataType{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		identifierTokens = append(identifierTokens, p.peek())
		identifiers = append(identifiers, p.peek().Value)
		variableTypes = append(variableTypes, variableType)

		// Check if variable is redeclared
		symbol := p.getSymbol(p.peek().Value)

		if symbol != nil {
			p.newError(p.peek().Position, fmt.Sprintf("Variable %s is redeclared in this scope.", p.consume().Value))
		} else {
			p.consume()
		}

		// More identifiers
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
		} else {
			break
		}
	}

	// Create node
	declareNode := &Node{startPosition.SetEndPos(identifierTokens[len(identifierTokens)-1].Position), NT_VariableDeclare, &VariableDeclareNode{variableType, constant, identifiers}}

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
		declareNode, expressionType = p.parseAssign(identifierTokens, variableTypes)

		// Change variable type if no was provided
		if variableType.DType == data.DT_NoType {
			variableType = expressionType
			node.Value.(*VariableDeclareNode).DataType = expressionType
		}
	}

	// Insert symbols
	for _, id := range identifiers {
		p.insertSymbol(id, &Symbol{ST_Variable, &VariableSymbol{variableType, declareNode.NodeType == NT_Assign, constant}})
	}

	return declareNode
}

func (p *Parser) parseAssign(identifierTokens []*lexer.Token, variableTypes []data.DataType) (*Node, data.DataType) {
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
		for i, identifier := range identifierTokens {
			// Variable doesn't have type
			if variableTypes[i].DType == data.DT_NoType {
				variableTypes[i] = expressionType
				// Variable has a type and it's incompatible with expression
			} else if !expressionType.Equals(variableTypes[i]) {

				// Assign type to empty list literal
				if variableTypes[i].DType == data.DT_List && expression.NodeType == NT_List && len(expression.Value.(*ListNode).Nodes) == 0 {
					expression.Value.(*ListNode).DataType.SubType = variableTypes[i].SubType
					continue
				}

				p.newErrorNoMessage()
				logger.Error2CodePos(identifierTokens[i].Position, &expressionPosition, fmt.Sprintf("Can't assign expression of type %s to variable %s of type %s.", expressionType, identifier, variableTypes[i]))
			}
		}
	}

	// Operation-Assign nodes
	if assign.TokenType != lexer.TT_KW_Assign {
		nodeType := OperationAssignTokenToNodeType[assign.TokenType]

		for i, identifier := range identifierTokens[:len(identifierTokens)-1] {
			// Transform a += 1 to a = a + 1
			variableNode := &Node{identifierTokens[i].Position, NT_Variable, &VariableNode{identifier.Value, expressionType}}
			newExpression := &Node{assign.Position, NT_Assign, &AssignNode{identifier.Value, &Node{assign.Position, nodeType, &BinaryNode{variableNode, expression, expressionType}}}}

			p.GetExpressionType(newExpression)
			visualize(newExpression, "", true)

			p.appendScope(newExpression)
		}

		// Transform a += 1 to a = a + 1
		variableNode := &Node{identifierTokens[len(identifierTokens)-1].Position, NT_Variable, &VariableNode{identifierTokens[len(identifierTokens)-1].Value, expressionType}}
		newExpression := &Node{assign.Position, nodeType, &BinaryNode{variableNode, expression, expressionType}}

		p.GetExpressionType(newExpression)

		return &Node{assign.Position, NT_Assign, &AssignNode{identifierTokens[len(identifierTokens)-1].Value, newExpression}}, expressionType
	}

	// Assign nodes
	for _, identifier := range identifierTokens[:len(identifierTokens)-1] {
		p.appendScope(&Node{assign.Position, NT_Assign, &AssignNode{identifier.Value, expression}})
	}

	return &Node{assign.Position, NT_Assign, &AssignNode{identifierTokens[len(identifierTokens)-1].Value, expression}}, expressionType
}
