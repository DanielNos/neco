package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseFunctionDeclaration() *Node {
	start := p.consume().Position
	identifierToken := p.consume()

	// Find function symbol
	function := p.functions[p.functionIndex]
	p.functionIndex++

	// Enter scope
	p.consume()
	p.enterScope()

	// Insert parameters to scope
	for _, parameter := range function.parameters {
		p.insertSymbol(parameter.Identifier, &Symbol{ST_Variable, &VariableSymbol{parameter.DataType, true, false}})
	}

	// Move to body
	var returnPosition *data.CodePos

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
	if function.returnType.Type != data.DT_Unknown {
		if !p.verifyReturns(body, function.returnType) {
			p.newError(returnPosition, "Function "+identifierToken.Value+" with return type "+function.returnType.String()+" does not return a value in all code paths.")
		}
	}

	p.leaveScope()

	// Store function name as a string constant for scope trace back
	p.StringConstants[identifierToken.Value] = -1
	return &Node{start.SetEndPos(p.peekPrevious().Position), NT_FunctionDeclaration, &FunctionDeclareNode{p.functionIndex - 1, identifierToken.Value, function.parameters, function.returnType, body}}
}

func (p *Parser) parseParameters() []Parameter {
	var paremeters = []Parameter{}

	// No parameters
	if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		return paremeters
	}

	for {
		// Collect data typa and identifier
		dataType := p.parseType()
		identifier := p.consume().Value

		// Create parameter and symbol
		paremeters = append(paremeters, Parameter{dataType, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{dataType, true, false}})

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()

		for p.peek().TokenType == lexer.TT_Identifier {
			// Create parameter and symbol
			identifier = p.consume().Value
			paremeters = append(paremeters, Parameter{dataType, identifier, nil})
			p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{dataType, true, false}})

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
	returnType := &data.DataType{data.DT_Unknown, nil}
	functionNumber := -1

	// Try to match arguments to some function from the bucket
	if functionBucketSymbol != nil && !errorsInArguments {
		// Function is picked by matching arguments
		functionSymbol := p.matchArguments(functionBucketSymbol, arguments, identifier)

		if functionSymbol != nil {
			// Store values from function symbol
			returnType = functionSymbol.returnType
			functionNumber = functionSymbol.number

			// Set function as used
			functionSymbol.everCalled = true

			// Insert empty string argument to printLine function without arguments
			if identifier.Value == "printLine" && len(arguments) == 0 {
				arguments = append(arguments, &Node{Position: identifier.Position, NodeType: NT_Literal, Value: &LiteralNode{data.DT_String, ""}})
				argumentTypes = append(argumentTypes, &data.DataType{data.DT_String, nil})
				p.StringConstants[""] = -1
			}
		}
	}
	p.consume()

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{functionNumber, identifier.Value, arguments, argumentTypes, returnType}}
}

func (p *Parser) matchArguments(bucket *Symbol, arguments []*Node, identifierToken *lexer.Token) *FunctionSymbol {
	// Collect argument data types
	argumentTypes := make([]*data.DataType, len(arguments))

	for i, argument := range arguments {
		argumentTypes[i] = GetExpressionType(argument)
	}

	for _, function := range bucket.value.(symbolTable) {
		// Incorrect argument amount
		if len(function.value.(*FunctionSymbol).parameters) != len(arguments) {
			continue
		}

		// Try to match arguments to parameters
		matched := true
		for i, parameter := range function.value.(*FunctionSymbol).parameters {
			if !parameter.DataType.CanBeAssigned(argumentTypes[i]) {
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
	p.newError(identifierToken.Position, "Failed to match function "+identifierToken.Value+" to any function header. Check if all arguments have the correct type and if there is the correct amount of them.")
	return nil
}

func (p *Parser) parseArguments() ([]*Node, []*data.DataType, bool) {
	arguments := []*Node{}
	argumentTypes := []*data.DataType{}

	// No arguments
	if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		return arguments, argumentTypes, false
	}

	// Collect arguments
	errorCount := p.ErrorCount
	var argument *Node
	for {
		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}

		argument = p.parseExpressionRoot()
		argumentTypes = append(argumentTypes, GetExpressionType(argument))
		arguments = append(arguments, argument)

		if p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
		}

		if p.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			break
		}
		p.consume()
	}

	return arguments, argumentTypes, errorCount != p.ErrorCount
}

func (p *Parser) verifyReturns(statementList *Node, returnType *data.DataType) bool {
	for _, statement := range statementList.Value.(*ScopeNode).Statements {
		// Return
		if statement.NodeType == NT_Return {
			// No return value
			if statement.Value == nil {
				p.newError(statement.Position, "Return statement has no return value, but function has return type "+returnType.String()+".")
			} else {
				// Incorrect return value data type
				expressionType := GetExpressionType(statement.Value.(*Node))

				if !returnType.CanBeAssigned(expressionType) {
					p.newError(statement.Value.(*Node).Position, "Return statement has return value with type "+expressionType.String()+", but function has return type "+returnType.String()+".")
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
