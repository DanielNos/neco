package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseIdentifierStatement() *Node {
	// Struct or enum variable declaration
	symbol := p.getGlobalSymbol(p.peek().Value)
	if symbol != nil && (symbol.symbolType == ST_Struct || symbol.symbolType == ST_Enum) {
		return p.parseVariableDeclaration(false)
	}

	// Collect statement expressions
	startPosition := p.peek().Position
	expressions := []*Node{p.parseIdentifier()}

	for p.peek().TokenType == lexer.TT_DL_Comma {
		p.consume()
		expressions = append(expressions, p.parseIdentifier())
	}

	// Statement is a function call
	if len(expressions) == 1 && expressions[0].NodeType == NT_FunctionCall && p.peek().TokenType == lexer.TT_EndOfCommand {
		return expressions[0]
	}

	// Missing assignment
	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.newError(startPosition.SetEndPos(p.peekPrevious().Position), "Expression list can't be a statement.")
		return nil
	}

	// Invalid tokens instead of assignment
	if !p.peek().TokenType.IsAssignKeyword() {
		startPosition = p.peek().Position
		for p.peek().TokenType != lexer.TT_EndOfCommand {
			p.consume()
		}

		p.newError(startPosition.SetEndPos(p.peekPrevious().Position), "Expected \"=\" after list of expressions.")
		return nil
	}

	// Check if all expressions are assignable
	for _, expression := range expressions {
		if expression.NodeType == NT_FunctionCall {
			p.newError(expression.Position, "Can't assign to a function call.")
		} else if expression.NodeType.IsOperator() {
			p.newError(expression.Position, "Can't assign to an expression.")
		}
	}

	node, _ := p.parseAssign(expressions, startPosition)
	return node
}

func (p *Parser) parseVariableIdentifiers(variableType data.DataType) ([]*Node, []string) {
	// Collect identifiers
	variables := []*Node{}
	identifiers := []string{}

	for p.peek().TokenType != lexer.TT_EndOfFile {
		identifiers = append(identifiers, p.peek().Value)
		variables = append(variables, &Node{p.peek().Position, NT_Variable, &VariableNode{p.peek().Value, variableType}})

		// Check if variable is redeclared
		symbol := p.getSymbol(p.peek().Value)

		if symbol != nil {
			p.newError(p.peek().Position, "Variable "+p.consume().Value+" is redeclared in this scope.")
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

	return variables, identifiers
}
