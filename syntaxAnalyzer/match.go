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
	hasDefault := false
	foundFirstCase := false

	for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_BraceClose {
		if sn.peek().TokenType == lexer.TT_KW_case {
			sn.consume() // case
			foundFirstCase = true

			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newError(sn.peekPrevious(), "Expected expression after case keyword.")
			} else {
				sn.analyzeExpression()
			}

			if sn.peek().TokenType != lexer.TT_DL_Colon {
				sn.newError(sn.peek(), "Expected colon after case expression.")
			} else {
				sn.consume()
			}

		} else if sn.peek().TokenType == lexer.TT_KW_default {
			if hasDefault {
				sn.newError(sn.peek(), "Multiple default cases arent't allowed.")
			}
			hasDefault = true

			sn.consume()
			if sn.peek().TokenType != lexer.TT_DL_Colon {
				sn.newError(sn.peek(), "Expected colon after keyword default.")
			} else {
				sn.consume()
			}

		} else {
			if !foundFirstCase && sn.peek().TokenType != lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), "Statement is outside of a case block.")
			}
			sn.analyzeStatement(false)
		}
	}
	sn.consume() // }
}
