package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeFunctionCall() {
	sn.consume() // (
	sn.analyzeArguments()

	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Check closing parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), "Expected \")\" after function call arguments, found \""+sn.peek().String()+"\" instead.")

		for sn.peek().TokenType != lexer.TT_EndOfCommand {
			sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" in function call.")
		}

		return
	}

	sn.consume() // )
}

func (sn *SyntaxAnalyzer) analyzeArguments() {
	for sn.peek().TokenType != lexer.TT_EndOfFile {
		if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			return
		}

		// Consume EOC
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}

		sn.analyzeExpression()

		// Next argument
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()
			// End of arguments
		} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose || sn.peek().TokenType == lexer.TT_EndOfCommand {
			return
			// Invalid token
		} else {
			sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" in argument.")
		}
	}
}
