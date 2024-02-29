package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeForEachLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword forEach, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check type
	if !sn.peek().TokenType.IsVariableType() {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check variable identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check keyword in
	if sn.peek().TokenType != lexer.TT_KW_in {
		sn.newError(sn.peek(), fmt.Sprintf("Expected keyword in after variable identifier, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check list identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check closing parenthesis
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.consume()
	} else {
		for sn.peek().TokenType != lexer.TT_DL_ParenthesisClose && sn.peek().TokenType != lexer.TT_DL_BraceOpen && sn.peek().TokenType != lexer.TT_EndOfCommand {
			sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis, found \"%s\" instead.", sn.consume()))
		}

		// Parenthesis found
		if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			sn.consume()
		}
	}

	// Check code block
	if sn.lookFor(lexer.TT_DL_BraceOpen, "forEach statement", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeForLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword for, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check init statement
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		// Missing init statement
		if sn.peek().Value == "" {
			sn.newError(sn.peek(), "For loop missing init statement.")
			return
		} else {
			sn.consume()
		}
		// No init statement
	} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.consume(), "For loop missing condition and step statement.")
		return
		// Check init statement
	} else {
		sn.analyzeStatement(false)

		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}
	}

	// Check condition
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		// Missing condition
		if sn.peek().Value == "" {
			sn.newError(sn.peek(), "For loop missing condition and step statement.")
			return
		} else {
			sn.consume()
		}
		// No condition
	} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.consume(), "For loop missing condition and step statement.")
		return
		// Check condition expression
	} else {
		sn.analyzeExpression()

		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}
	}

	// Empty step
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
		// No step
	} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.consume()
		// Check step
	} else {
		sn.analyzeStatement(false)
	}

	if sn.lookFor(lexer.TT_DL_BraceOpen, "for loop header", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeWhileLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword while, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check condition
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.newError(sn.peek(), fmt.Sprintf("Expected condition, found \"%s\" instead.", sn.peek()))
		return
	}
	sn.analyzeExpression()

	// Check closing parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis after condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	if sn.lookFor(lexer.TT_DL_BraceOpen, "while loop condition", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeLoop() {
	sn.consume()

	if sn.lookFor(lexer.TT_DL_BraceOpen, "keyword loop", "opening brace", false) {
		sn.analyzeScope()
	}
}
