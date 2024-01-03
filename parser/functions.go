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

	for p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		// Collect data typa and identifier
		dataType := TokenTypeToDataType[p.consume().TokenType]
		identifier := p.consume().Value

		// Create parameter and symbol
		paremeters = append(paremeters, Parameter{dataType, identifier, nil})
		p.insertSymbol(identifier, &Symbol{ST_Variable, &VariableSymbol{VariableType{dataType, false}, true, false}})
	}

	return paremeters
}

func (p *Parser) parseFunctionCall(functionSymbol *Symbol, identifier *lexer.Token) *Node {
	// Collect arguments
	arguments := p.parseArguments(functionSymbol.value.(*FunctionSymbol).parameters)

	return &Node{identifier.Position, NT_FunctionCall, &FunctionCallNode{identifier.Value, arguments, &functionSymbol.value.(*FunctionSymbol).returnType}}
}

func (p *Parser) parseArguments(paramters []Parameter) []*Node {
	p.consume()
	var arguments = []*Node{}

	// Collect arguments
	for p.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		arguments = append(arguments, p.parseExpression(MINIMAL_PRECEDENCE))
	}
	p.consume()

	return arguments
}

func (p *Parser) verifyReturns(statementList *Node, returnType VariableType) bool {	
	for _, statement := range statementList.value.(*ScopeNode).statements {
		// Return
		if statement.nodeType == NT_Return {
			// No return value
			if statement.value == nil {
				p.newError(statement.position, fmt.Sprintf("Return statement has no return value, but function has return type %s.", returnType))
			} else {
				// Incorrect return value data type
				expressionType := p.getExpressionType(statement.value.(*Node))
				
				if !returnType.Equals(expressionType) {
					position := getExpressionPosition(statement.value.(*Node), statement.value.(*Node).position.StartChar, statement.value.(*Node).position.EndChar)
					p.newError(&position, fmt.Sprintf("Return statement has return value with type %s, but function has return type %s.", expressionType, returnType))
				}
			}

			return true
		}
	}

	return false
}
