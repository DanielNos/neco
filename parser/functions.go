package parser

import (
	"fmt"
	"neko/dataStructures"
	"neko/lexer"
)

func (p *Parser) parseFunctionDeclare() *Node {
	start := p.consume().Position

	// Check for redeclaration
	symbol := p.findSymbol(p.peek().Value)

	if symbol != nil {
		p.newError(p.peek().Position, fmt.Sprintf("Redeclaration of symbol %s.", p.peek().Value))
	}

	// Collect name
	identifier := p.consume().Value

	// Enter scope
	p.enterScope()

	p.consume()

	// Collect parameters
	parameters := p.parseParameters()

	p.consume()

	// Collect return type
	returnType := VariableType{DT_NoType, false}
	var returnPosition *dataStructures.CodePos

	if p.peek().TokenType == lexer.TT_KW_returns {
		returnPosition = p.consume().Position
		returnType.dataType = TokenTypeToDataType[p.consume().TokenType]
		returnPosition.EndChar = p.peekPrevious().Position.EndChar
	}

	// Parse body
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}
	body := &Node{p.peek().Position, NT_Scope, p.parseScope(false)}

	// Check if function has return statements in all paths
	if returnType.dataType != DT_NoType {
		if !p.verifyReturns(body, returnType) {
			p.newError(returnPosition, fmt.Sprintf("Function %s with return type %s does not return a value in all code paths.", identifier, returnType))
		}
	}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.symbolTableStack.Pop()

	// Insert function symbol
	p.insertSymbol(identifier, &Symbol{ST_Function, &FunctionSymbol{parameters, returnType}})

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

func (p *Parser) parseFunctionCall(functionSymbol *Symbol, identifier *lexer.Token) *Node {
	// Collect arguments
	var parameters *[]Parameter = nil
	returnType := &VariableType{DT_NoType, false}

	if functionSymbol != nil {
		parameters = &functionSymbol.value.(*FunctionSymbol).parameters
		returnType = &functionSymbol.value.(*FunctionSymbol).returnType
	}

	arguments := p.parseArguments(parameters, identifier.Value, identifier.Position)
	p.consume()

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{identifier.Value, arguments, returnType}}
}

func (p *Parser) parseArguments(parameters *[]Parameter, functionName string, functionPosition *dataStructures.CodePos) []*Node {
	p.consume()
	var arguments = []*Node{}

	// No parameters, collect arguments any arguments
	if parameters == nil {
		p.parseAnyArguments()
		// Check arguments
	} else {
		// No arguments
		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			p.newError(functionPosition, fmt.Sprintf("Function %s has %d parameters, but was called with no arguments.", functionName, len(*parameters)))
			return arguments
		}

		// Check if arguments have same type as parameters
		for parameterIndex := 0; parameterIndex < len(*parameters); parameterIndex++ {
			// Collect argument
			argument := p.parseExpression(MINIMAL_PRECEDENCE)
			argumentType := p.getExpressionType(argument)

			// Check type
			if !argumentType.Equals((*parameters)[parameterIndex].DataType) {
				argumentPosition := getExpressionPosition(argument, argument.Position.StartChar, argument.Position.EndChar)
				p.newError(&argumentPosition, fmt.Sprintf("Function's %s argument \"%s\" has type %s, but it should be %s.", functionName, (*parameters)[parameterIndex].Identifier, argumentType, (*parameters)[parameterIndex].DataType))
			}

			// Store argument
			arguments = append(arguments, argument)

			if parameterIndex != len(*parameters)-1 {
				if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
					p.newError(functionPosition, fmt.Sprintf("Function %s has %d parameter/s, but was called with only %d argument/s.", functionName, len(*parameters), parameterIndex+1))
					return arguments
				}
				p.consume()
			}
		}

		// More arguments than parameters
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
			argumentCount := p.parseAnyArguments()
			p.newError(functionPosition, fmt.Sprintf("Function %s has %d parameter/s, but was called with %d argument/s.", functionName, len(*parameters), len(*parameters)+argumentCount))
		}
	}

	return arguments
}

func (p *Parser) parseAnyArguments() int {
	argumentCount := 0

	for {
		p.parseExpression(MINIMAL_PRECEDENCE)
		argumentCount++

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()
	}

	return argumentCount
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
