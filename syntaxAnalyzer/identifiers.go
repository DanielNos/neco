package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeIdentifier() {
	// Function call
	if sn.peekNext().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeIdentifierStatement()
		// Variable
	} else {
		// Assignment
		if sn.peekNext().TokenType.IsAssignKeyword() {
			sn.consume()
			sn.consume()
			sn.analyzeExpression()
			return
		}
		// Assignment to multiple variables
		if sn.peekNext().TokenType == lexer.TT_DL_Comma {
			sn.consume()
			for sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.consume()
				if sn.peek().TokenType != lexer.TT_Identifier {
					sn.newError(sn.peek(), "Expected variable identifier after comma, found "+sn.peek().String()+" instead.")
					break
				}
				sn.consume()
			}

			// No = after variables
			if !sn.peek().TokenType.IsAssignKeyword() {
				sn.newError(sn.peek(), "Expected = after list of variable identifiers, found "+sn.peek().String()+" instead.")
				return
			}

			// Assignment
			sn.consume()
			sn.analyzeExpression()
			return
		}

		// Declare custom variable
		if sn.customTypes[sn.peek().Value] {
			sn.consume()

			// Check identifier
			if sn.peek().TokenType != lexer.TT_Identifier {
				sn.newError(sn.peek(), "Expected variable identifier after type "+sn.peekPrevious().Value+", found \""+sn.peek().String()+"\" instead.")
			} else {
				sn.consume()
			}

			// Assign to it
			if sn.peek().TokenType == lexer.TT_KW_Assign {
				sn.consume()
				sn.analyzeExpression()
			}
			return
		}

		// Assigning to list index
		if sn.peekNext().TokenType == lexer.TT_DL_BracketOpen {
			startChar := sn.consume().Position.StartChar
			openingBracket := sn.consume() // Collect [

			// Analyze index expression
			sn.analyzeExpression()

			// Missing closing bracket
			if sn.peek().TokenType != lexer.TT_DL_BracketClose {
				sn.newError(openingBracket, "Index is missing closing bracket.")
				return
			}

			sn.consume() // Collect ]

			// End of expression
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newErrorFromTo(sn.peek().Position.StartLine, startChar, sn.peek().Position.StartChar, "Expression can't be a statement.")
				return
			}

			// No assignment
			if !sn.peek().TokenType.IsAssignKeyword() {
				return
			}

			// Assign expression
			sn.consume() // Collect =
			sn.analyzeExpression()

			return
		}

		// Expression
		startChar := sn.peek().Position.StartChar
		sn.analyzeExpression()
		sn.newErrorFromTo(sn.peek().Position.StartLine, startChar, sn.peek().Position.StartChar, "Expression can't be a statement.")
	}
}

func (sn *SyntaxAnalyzer) analyzeIdentifierStatement() {
	// Enum variable declaration
	if sn.customTypes[sn.peek().Value] {
		sn.analyzeVariableDeclaration(false)
		return
	}

	sn.consume()

	// Function call
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeFunctionCall()
		// Assignment
	} else if sn.peek().TokenType.IsAssignKeyword() {
		sn.consume() // =
		sn.analyzeExpression()
	}
}
