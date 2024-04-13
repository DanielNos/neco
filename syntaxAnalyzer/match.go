package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeMatchStatement() {
	sn.consume() // match

	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.newError(sn.peek(), "Expected matched expression after match keyword.")
	} else {
		sn.analyzeExpression()
	}

	if !sn.lookFor(lexer.TT_DL_BraceOpen, "matched expression", "opening brace", false) {
		return
	}

	sn.consume() // {
	foundFirstCase := false

	for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_BraceClose {
		// Case
		if sn.peek().TokenType == lexer.TT_KW_case {
			sn.consume() // case
			foundFirstCase = true

			// Analyze expression
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newError(sn.peekPrevious(), "Expected expression after case keyword.")
			} else {
				sn.analyzeExpression()
			}

			// More expressions
			for sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.consume()
				sn.analyzeExpression()
			}

			// Check for colon
			if sn.peek().TokenType != lexer.TT_DL_Colon {
				sn.newError(sn.peek(), "Expected colon after case expression.")
			} else {
				sn.consume()
			}

			// Default
		} else if sn.peek().TokenType == lexer.TT_KW_default {
			sn.consume()
			if sn.peek().TokenType != lexer.TT_DL_Colon {
				sn.newError(sn.peek(), "Expected colon after keyword default.")
			} else {
				sn.consume()
			}

			// Statements in cases
		} else {
			if !foundFirstCase && sn.peek().TokenType != lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), "Statement is outside of a case block.")
			}
			sn.analyzeStatement(false)
		}
	}
	sn.consume() // }
}
