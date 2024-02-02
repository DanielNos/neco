package parser

import (
	"fmt"
	"neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseFunctionDeclare() *Node {
	start := p.consume().Position

	// Find bucket
	identifier := p.consume().Value
	symbol := p.findSymbol(identifier)

	// Enter scope
	p.enterScope()

	p.consume()

	// Collect parameters
	parameters := p.parseParameters()

	// Check for redeclaration
	if symbol != nil {
		// Create paramters id and look for a function using
		if symbol.symbolType == ST_FunctionBucket {
			id := createParametersIdentifier(parameters)
			if symbol.value.(symbolTable)[id] != nil {
				p.newError(p.peek().Position, fmt.Sprintf("Redeclaration of symbol %s.", p.peek().Value))
			}
		}
	}

	p.consume()

	// Collect return type
	returnType := VariableType{DT_NoType, false}
	var returnPosition *dataStructures.CodePos

	if p.peek().TokenType == lexer.TT_KW_returns {
		returnPosition = p.consume().Position
		returnType.DataType = TokenTypeToDataType[p.consume().TokenType]
		returnPosition.EndChar = p.peekPrevious().Position.EndChar
	}

	// Parse body
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}
	body := p.parseScope(false, true).(*Node)

	// Check if function has return statements in all paths
	if returnType.DataType != DT_NoType {
		if !p.verifyReturns(body, returnType) {
			p.newError(returnPosition, fmt.Sprintf("Function %s with return type %s does not return a value in all code paths.", identifier, returnType))
		}
	}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.symbolTableStack.Pop()

	// Insert function symbol
	p.insertFunction(identifier, parameters, returnType)

	return &Node{start, NT_FunctionDeclare, &FunctionDeclareNode{identifier, parameters, returnType, body}}
}

func (p *Parser) parseParameters() []Parameter {
	var paremeters = []Parameter{}

	if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		return paremeters
	}

	for {
		// Collect data typa and identifier
		dataType := TokenTypeToDataType[p.consume().TokenType]
		identifier := p.consume().Value

		// Create parameter and symbol
		paremeters = append(paremeters, Parameter{VariableType{dataType, false}, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{VariableType{dataType, false}, true, false}})

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()

		for p.peek().TokenType == lexer.TT_Identifier {
			// Create parameter and symbol
			identifier = p.consume().Value
			paremeters = append(paremeters, Parameter{VariableType{dataType, false}, identifier, nil})
			p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{VariableType{dataType, false}, true, false}})

			p.consume()
		}
	}

	return paremeters
}

func (p *Parser) parseFunctionCall(functionBucketSymbol *Symbol, identifier *lexer.Token) *Node {
	// Collect arguments
	arguments := p.parseArguments()

	// Check if arguments match any function
	if functionBucketSymbol != nil {
		p.matchArguments(functionBucketSymbol, arguments, identifier)
	}
	p.consume()

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{identifier.Value, arguments, &VariableType{DT_NoType, false}}}
}

func (p *Parser) matchArguments(bucket *Symbol, arguments []*Node, identifierToken *lexer.Token) {
	// Collect argument data types
	argumentTypes := make([]VariableType, len(arguments))

	for i, argument := range arguments {
		argumentTypes[i] = p.getExpressionType(argument)
	}

	for _, function := range bucket.value.(map[string]*Symbol) {
		// Incorrect argument amount
		if len(function.value.(*FunctionSymbol).parameters) != len(arguments) {
			continue
		}

		// Try to match arguments to paramters
		matched := true
		for i, parameter := range function.value.(*FunctionSymbol).parameters {
			if !parameter.DataType.Equals(argumentTypes[i]) {
				matched = false
				break
			}
		}

		// Failed to match
		if !matched {
			continue
		}

		// Successfully matched
		return
	}

	// Failed to match to all functions in a bucket
	p.newError(identifierToken.Position, fmt.Sprintf("Failed to match function %s to any header.", identifierToken.Value))
}

func (p *Parser) parseArguments() []*Node {
	arguments := []*Node{}

	for {
		arguments = append(arguments, p.parseExpressionRoot())

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()
	}

	return arguments
}

func (p *Parser) verifyReturns(statementList *Node, returnType VariableType) bool {
	for _, statement := range statementList.Value.(*ScopeNode).Statements {
		// Return
		if statement.NodeType == NT_Return {
			// No return value
			if statement.Value == nil {
				p.newError(statement.Position, fmt.Sprintf("Return statement has no return value, but function has return type %s.", returnType))
			} else {
				// Incorrect return value data type
				expressionType := p.getExpressionType(statement.Value.(*Node))

				if !returnType.Equals(expressionType) {
					position := getExpressionPosition(statement.Value.(*Node), statement.Value.(*Node).Position.StartChar, statement.Value.(*Node).Position.EndChar)
					p.newError(&position, fmt.Sprintf("Return statement has return value with type %s, but function has return type %s.", expressionType, returnType))
				}
			}

			return true
			// If statement
		} else if statement.NodeType == NT_If {
			ifNode := statement.Value.(*IfNode)

			// Check if body
			if !p.verifyReturns(ifNode.Body, returnType) {
				return false
			}

			// Check else body
			if ifNode.ElseBody != nil && !p.verifyReturns(ifNode.ElseBody, returnType) {
				return false
			}

			// Check else if bodies
			for _, elif := range ifNode.ElseIfs {
				if !p.verifyReturns(elif.Value.(*IfNode).Body, returnType) {
					return false
				}
			}

			return true
		}
	}

	return false
}
