package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeMatchStatement(isExpression bool) {
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

	for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_BraceClose {
		// Skip empty lines
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
			continue
		}

		// Case
		if sn.peek().TokenType.CanBeExpression() {
			// Analyze expression
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newError(sn.peekPrevious(), "Expected case expression.")
			} else {
				sn.analyzeExpression()
			}

			// More expressions
			for sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.consume()
				sn.analyzeExpression()
			}

			// Check for colon
			if sn.peek().TokenType != lexer.TT_KW_CaseIs {
				sn.newError(sn.peek(), "Expected \"=>\" after case expression.")
			} else {
				sn.consume()
			}

			sn.consumeEOCs()

			if isExpression {
				sn.analyzeExpression()
			} else {
				sn.analyzeStatement(false)
			}

			// Default
		} else if sn.peek().TokenType == lexer.TT_KW_default {
			sn.consume()
			if sn.peek().TokenType != lexer.TT_KW_CaseIs {
				sn.newError(sn.peek(), "Expected \"=>\" after keyword default.")
			} else {
				sn.consume()
			}

			sn.consumeEOCs()

			if isExpression {
				sn.analyzeExpression()
			} else {
				sn.analyzeStatement(false)
			}

			// Statements/Expressions outside of cases
		} else {
			if isExpression {
				sn.newError(sn.consume(), "Expression is outside of a case block.")
			} else {
				sn.newError(sn.consume(), "Statement is outside of a case block.")
			}
		}
	}
	sn.consume() // }
}
