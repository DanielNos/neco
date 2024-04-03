package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
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
		if variableType.Type == data.DT_Unknown {
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
		var expressionType *data.DataType
		declareNode, expressionType = p.parseAssignment(variableNodes, startPosition)

		// Change variable type if no was provided
		if variableType.Type == data.DT_Unknown {
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

func (p *Parser) parseAssignment(assignedTo []*Node, startOfStatement *data.CodePos) (*Node, *data.DataType) {
	assign := p.consume()
	expressionStart := p.peek().Position

	// Collect expression
	expression := p.parseExpressionRoot()
	expressionType := p.GetExpressionType(expression)

	// Uncompatible data types
	expressionPosition := data.CodePos{expressionStart.File, expressionStart.StartLine, expressionStart.EndLine, expressionStart.StartChar, p.peekPrevious().Position.EndChar}

	// Check if variables are constants
	for _, target := range assignedTo {
		symbol := p.findSymbol(target.Value.(*VariableNode).Identifier)

		if symbol == nil || symbol.symbolType != ST_Variable {
			continue
		}

		if symbol.value.(*VariableSymbol).isConstant {
			p.newError(GetExpressionPosition(target), "Variable "+target.Value.(*VariableNode).Identifier+" is constant.")
		}
	}

	// Print errors
	if expressionType.Type != data.DT_Unknown {
		for _, target := range assignedTo {
			targetType := p.GetExpressionType(target)

			if targetType.Type == data.DT_Unknown {
				// Sub-type can be determined, assign it to expression
				if expressionType.IsComplete() {
					targetType = expressionType
				} else {
					p.newError(&expressionPosition, "Can't assign expression with type "+expressionType.String()+" to a variable declared using var, because sub-type can't be determined. Replace var with desired type or add type before expression.")
				}
				continue
			}

			// Can't be assgined
			if !targetType.CanBeAssigned(expressionType) {
				// Type is complete
				if expressionType.IsComplete() {
					p.newError(&expressionPosition, "Cant't assign expression with type "+expressionType.String()+" to variable with type "+targetType.String()+".")
					continue
				}

				// Type doesn't have a leaf type, set it to the same as target
				originalExpressionType := expressionType.String()
				expressionTypeCopy := expressionType.Copy()
				expressionTypeCopy.TryCompleteFrom(targetType)

				// Check if now it can be assigned
				if !targetType.CanBeAssigned(expressionTypeCopy) {
					p.newError(&expressionPosition, "Cant't assign expression with type "+originalExpressionType+" to variable with type "+targetType.String()+".")
				}
			}
		}
	}

	// Operation-assign nodes need to be edited
	if assign.TokenType != lexer.TT_KW_Assign {
		nodeType := OperationAssignTokenToNodeType[assign.TokenType]

		// Transform assigned expressions in the following way: a += 1 to a = a + 1
		for _, target := range assignedTo[:len(assignedTo)-1] {
			generatedNode := &Node{assign.Position, nodeType, &TypedBinaryNode{target, expression, expressionType}}
			p.appendScope(&Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{[]*Node{target}, generatedNode}})
		}

		generatedNode := &Node{assign.Position, nodeType, &TypedBinaryNode{assignedTo[len(assignedTo)-1], expression, expressionType}}
		return &Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{[]*Node{assignedTo[len(assignedTo)-1]}, generatedNode}}, expressionType
	}

	return &Node{startOfStatement.SetEndPos(p.peekPrevious().Position), NT_Assign, &AssignNode{assignedTo, expression}}, expressionType
}
