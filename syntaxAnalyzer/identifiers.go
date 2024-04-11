package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeIdentifierStatement() {
	// Enum variable declaration
	if sn.customTypes[sn.peek().Value] {
		sn.analyzeVariableDeclaration(false)
		return
	}

	// Function call
	if sn.peekNext().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.consume() // (
		sn.analyzeFunctionCall()
		return
	}

	// Assignment
	startOfStatement := sn.peek()
	sn.analyzeIdentifier()

	// Multiple identifiers
	for sn.peek().TokenType == lexer.TT_DL_Comma {
		sn.consume()

		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.newError(sn.peek(), "Expected identifier after comma, found EOC instead.")
			return
		}

		sn.analyzeIdentifier()
	}

	// Analyze assignment
	if sn.peek().TokenType.IsAssignKeyword() {
		sn.analyzeAssignment()
	} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
		// Missing assign keyword
		sn.newErrorFromTo(sn.peek().Position.StartLine, startOfStatement.Position.StartChar, sn.peekPrevious().Position.EndChar, "Expression can't be a statement.")
	} else if sn.peek().TokenType.IsOperator() {
		// Operator after identifiers
		sn.analyzeExpression()
		sn.newErrorFromTo(sn.peek().Position.StartLine, startOfStatement.Position.StartChar, sn.peekPrevious().Position.EndChar, "Expression can't be a statement.")
	} else {
		// Tokens after identifiers
		for sn.peek().TokenType != lexer.TT_EndOfCommand {
			sn.consume()
		}

		sn.newErrorFromTo(sn.peek().Position.StartLine, startOfStatement.Position.StartChar, sn.peekPrevious().Position.EndChar, "Invalid statement.")
	}
}

func (sn *SyntaxAnalyzer) analyzeIdentifier() {
	sn.consume() // ID

	// List element
	if sn.peek().TokenType == lexer.TT_DL_BracketOpen {
		openingBracket := sn.consume() // [

		sn.analyzeExpression()

		// Missing closing bracket
		if sn.peek().TokenType != lexer.TT_DL_BracketClose {
			sn.newError(openingBracket, "Index is missing closing bracket.")
			return
		}

		sn.consume() // ]
	}

	// Object field
	if sn.peek().TokenType == lexer.TT_OP_Dot {
		sn.consume() // .
		sn.analyzeIdentifier()
	}

	// Function call
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeFunctionCall()
	}
}

func (sn *SyntaxAnalyzer) analyzeAssignment() {
	assignToken := sn.consume() // =

	// Missing expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.newError(sn.peek(), "Expected assigned expression after "+assignToken.String()+".")
		return
	}

	sn.analyzeExpression()
}
