package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeIfStatement(isElseIf bool) {
	sn.consume()

	statementName := "if"

	if isElseIf {
		statementName = "elif"
	}

	// Check opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword if, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Collect expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_EndOfFile {
		sn.newError(sn.peek(), fmt.Sprintf("Expected condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.analyzeExpression()
	}

	// Check closing parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis after condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(lexer.TT_DL_BraceOpen, "if statement", "opening brace", false) {
		return
	}

	// Check body
	sn.analyzeScope()

	// Check else statement
	if sn.peek().TokenType == lexer.TT_KW_else {
		sn.analyzeElseStatement()
		return
		// Check elif statement
	} else if sn.peek().TokenType == lexer.TT_KW_elif {
		sn.analyzeIfStatement(true)
		return
		// Skip 1 EOC
	} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()

		// Check else statement
		if sn.peek().TokenType == lexer.TT_KW_else {
			sn.consume()
			sn.analyzeElseStatement()
			return
			// Check elif statement
		} else if sn.peek().TokenType == lexer.TT_KW_elif {
			sn.analyzeIfStatement(true)
			return
		}
	}

	// Look for else or elif after many EOCs
	if sn.peekNext().TokenType != lexer.TT_EndOfCommand {
		return
	}

	for sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()

		// Found else
		if sn.peekNext().TokenType == lexer.TT_KW_else {
			sn.consume()
			sn.newError(sn.peek(), fmt.Sprintf("Too many EOCs (\\n or ;) after %s block. Only 0 or 1 EOCs are allowed.", statementName))
			sn.analyzeElseStatement()
			return
			// Found elif
		} else if sn.peekNext().TokenType == lexer.TT_KW_elif {
			sn.consume()
			sn.newError(sn.peek(), fmt.Sprintf("Too many EOCs (\\n or ;) after %s block. Only 0 or 1 EOCs are allowed.", statementName))
			sn.analyzeIfStatement(true)
			return
			// Other tokens
		} else if sn.peekNext().TokenType != lexer.TT_EndOfCommand {
			return
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeElseStatement() {
	sn.consume()

	// Check opening brace
	if sn.lookFor(lexer.TT_DL_BraceOpen, "else statement", "opening brace", false) {
		sn.analyzeScope()
	}
}
