package parser

import (
	"fmt"
	"neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseFunctionDeclare() *Node {
	start := p.consume().Position
	identifierToken := p.consume()

	// Find function symbol
	function := p.functions[p.functionIndex]
	p.functionIndex++

	// Enter scope
	p.enterScope()
	p.consume()

	// Move to body
	var returnPosition *dataStructures.CodePos

	for p.peek().TokenType != lexer.TT_EndOfCommand && p.peek().TokenType != lexer.TT_DL_BraceOpen {
		if p.peek().TokenType == lexer.TT_KW_returns {
			returnPosition = p.peek().Position
		}
		p.consume()
	}

	// Parse body
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}
	body := p.parseScope(false, true).(*Node)

	// Check if function has return statements in all paths
	if function.returnType.DType != DT_NoType {
		if !p.verifyReturns(body, function.returnType) {
			p.newError(returnPosition, fmt.Sprintf("Function %s with return type %s does not return a value in all code paths.", identifierToken.Value, function.returnType))
		}
	}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.stack_symbolTablestack.Pop()

	p.StringConstants[identifierToken.Value] = -1
	return &Node{start, NT_FunctionDeclare, &FunctionDeclareNode{p.functionIndex - 1, identifierToken.Value, function.parameters, function.returnType, body}}
}

func (p *Parser) parseFunctionHeader() {
	// Find bucket
	identifierToken := p.consume()
	symbol := p.findSymbol(identifierToken.Value)

	// Enter scope
	p.enterScope()
	p.consume()

	// Collect parameters
	parameters := p.parseParameters()

	// Function entry() can't have parameters
	if identifierToken.Value == "entry" && len(parameters) != 0 {
		// TODO: Display position of parameters
		p.newError(identifierToken.Position, "Function entry() can't have any parameters.")
	}

	// Check for redeclaration
	if symbol != nil {
		// Redeclaration of entry()
		if identifierToken.Value == "entry" {
			p.newError(identifierToken.Position, "Function entry() can't be overloaded.")
		}

		// Create parameters id and look for a function using it
		if symbol.symbolType == ST_FunctionBucket {
			id := createParametersIdentifier(parameters)
			if symbol.value.(symbolTable)[id] != nil {
				p.newError(identifierToken.Position, fmt.Sprintf("Redeclaration of symbol %s.", identifierToken.Value))
			}
		}
	}

	p.consume()

	// Collect return type
	returnType := DataType{DT_NoType, nil}
	var returnPosition *dataStructures.CodePos

	if p.peek().TokenType == lexer.TT_KW_returns {
		returnPosition = p.consume().Position
		returnType.DType = TokenTypeToDataType[p.consume().TokenType]
		returnPosition.EndChar = p.peekPrevious().Position.EndChar

		// Function entry() can't have a return type
		if identifierToken.Value == "entry" {
			p.newError(returnPosition, "Function entry() can't have a return type.")
		}
	}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.stack_symbolTablestack.Pop()

	// Insert function symbol
	newSymbol := p.insertFunction(identifierToken.Value, &FunctionSymbol{len(p.functions), parameters, returnType, identifierToken.Value == "entry"})
	p.functions = append(p.functions, newSymbol.value.(*FunctionSymbol))
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
		paremeters = append(paremeters, Parameter{DataType{dataType, nil}, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{DataType{dataType, nil}, true, false}})

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()

		for p.peek().TokenType == lexer.TT_Identifier {
			// Create parameter and symbol
			identifier = p.consume().Value
			paremeters = append(paremeters, Parameter{DataType{dataType, nil}, identifier, nil})
			p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{DataType{dataType, nil}, true, false}})

			p.consume()
		}
	}

	return paremeters
}

func (p *Parser) parseFunctionCall(functionBucketSymbol *Symbol, identifier *lexer.Token) *Node {
	// Collect arguments
	p.consume()
	arguments, argumentTypes, errorsInArguments := p.parseArguments()

	// Check if arguments match any function
	returnType := &DataType{DT_NoType, nil}
	functionNumber := -1

	// Try to match arguments to some function from the bucket
	if functionBucketSymbol != nil && !errorsInArguments {
		// Function is picked by matching arguments
		functionSymbol := p.matchArguments(functionBucketSymbol, arguments, identifier)

		if functionSymbol != nil {
			// Store values from function symbol
			returnType = &functionSymbol.returnType
			functionNumber = functionSymbol.number

			// Set function as used
			functionSymbol.everCalled = true
		}
	}
	p.consume()

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{functionNumber, identifier.Value, arguments, argumentTypes, returnType}}
}

func (p *Parser) matchArguments(bucket *Symbol, arguments []*Node, identifierToken *lexer.Token) *FunctionSymbol {
	// Collect argument data types
	argumentTypes := make([]DataType, len(arguments))

	for i, argument := range arguments {
		argumentTypes[i] = p.GetExpressionType(argument)
	}

	for _, function := range bucket.value.(symbolTable) {
		// Incorrect argument amount
		if len(function.value.(*FunctionSymbol).parameters) != len(arguments) {
			continue
		}

		// Try to match arguments to parameters
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
		return function.value.(*FunctionSymbol)
	}

	// Failed to match to all functions in a bucket
	p.newError(identifierToken.Position, fmt.Sprintf("Failed to match function %s to any header.", identifierToken.Value))
	return nil
}

func (p *Parser) parseArguments() ([]*Node, []DataType, bool) {
	arguments := []*Node{}
	argumentTypes := []DataType{}

	// No arguments
	if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		return arguments, argumentTypes, false
	}

	// Collect arguments
	errorCount := p.ErrorCount
	var argument *Node
	for {
		argument = p.parseExpressionRoot()
		argumentTypes = append(argumentTypes, p.GetExpressionType(argument))
		arguments = append(arguments, argument)

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()
	}

	return arguments, argumentTypes, errorCount != p.ErrorCount
}

func (p *Parser) verifyReturns(statementList *Node, returnType DataType) bool {
	for _, statement := range statementList.Value.(*ScopeNode).Statements {
		// Return
		if statement.NodeType == NT_Return {
			// No return value
			if statement.Value == nil {
				p.newError(statement.Position, fmt.Sprintf("Return statement has no return value, but function has return type %s.", returnType))
			} else {
				// Incorrect return value data type
				expressionType := p.GetExpressionType(statement.Value.(*Node))

				if !returnType.Equals(expressionType) {
					position := getExpressionPosition(statement.Value.(*Node), statement.Value.(*Node).Position.StartChar, statement.Value.(*Node).Position.EndChar)
					p.newError(position, fmt.Sprintf("Return statement has return value with type %s, but function has return type %s.", expressionType, returnType))
				}
			}

			return true
			// If statement
		} else if statement.NodeType == NT_If {
			ifNode := statement.Value.(*IfNode)

			// Check if bodies
			for _, ifStatement := range ifNode.IfStatements {
				if !p.verifyReturns(ifStatement.Body, returnType) {
					return false
				}
			}

			// Check else body
			if ifNode.ElseBody != nil && !p.verifyReturns(ifNode.ElseBody, returnType) {
				return false
			}

			return true
		}
	}

	return false
}
