package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeFunctionCall() {
	sn.consume()
	sn.analyzeAruments()

	// Check closing parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected \")\" after function call arguments, found \"%s\" instead.", sn.peek()))

		for sn.peek().TokenType != lexer.TT_EndOfCommand {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in function call.", sn.consume()))
		}

		return
	}
	sn.consume()
}

func (sn *SyntaxAnalyzer) analyzeAruments() {
	for sn.peek().TokenType != lexer.TT_EndOfFile {
		if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose || sn.peek().TokenType == lexer.TT_EndOfCommand {
			return
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
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in argument.", sn.consume()))
		}
	}
}
