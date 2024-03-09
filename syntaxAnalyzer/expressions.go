package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeExpression() {
	// No expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Operator
	if sn.peek().TokenType.IsUnaryOperator() {
		sn.consume()
		sn.analyzeExpression()
		return
	}
	// Literal
	if sn.peek().TokenType.IsLiteral() {
		sn.consume()

		// Not end of expression
		if sn.peek().TokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
		}
		return
	}
	// Variable or function call
	if sn.peek().TokenType == lexer.TT_Identifier {
		sn.consume()

		// Variable
		if sn.peek().TokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
			// Function call
		} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
			sn.analyzeFunctionCall()

			// Not end of expression
			if sn.peek().TokenType.IsBinaryOperator() {
				sn.consume()
				sn.analyzeExpression()
			}
			// List index
		} else if sn.peek().TokenType == lexer.TT_DL_BracketOpen {
			for sn.peek().TokenType == lexer.TT_DL_BracketOpen {
				sn.consume()

				// Index expression
				if sn.peek().TokenType == lexer.TT_DL_BracketClose {
					sn.newError(sn.peek(), "Expected list index.")
					return
				} else {
					sn.analyzeExpression()
				}

				// Closing bracket
				if sn.peek().TokenType != lexer.TT_DL_BracketClose {
					sn.newError(sn.peek(), "Expected closing bracket.")
				} else {
					sn.consume()
				}
			}
			// Struct
		} else if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.consume()

			for sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.consume()
			}

			for sn.peek().TokenType != lexer.TT_EndOfCommand {
				if sn.peek().TokenType == lexer.TT_DL_BraceClose {
					sn.consume()
					break
				}

				sn.analyzeExpression()

				if sn.peek().TokenType == lexer.TT_DL_BraceClose {
					sn.consume()
					break
				} else if sn.peek().TokenType == lexer.TT_DL_Comma {
					sn.consume()
				}

				for sn.peek().TokenType == lexer.TT_EndOfCommand {
					sn.consume()
				}
			}
		}

		return
	}
	// Sub-Expression
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeSubExpression()

		// Not end of expression
		if sn.peek().TokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
		}
		return
	}
	// Function call of function with same name as keyword
	if (sn.peek().TokenType == lexer.TT_KW_int || sn.peek().TokenType == lexer.TT_KW_flt || sn.peek().TokenType == lexer.TT_KW_str) && sn.peekNext().TokenType == lexer.TT_DL_ParenthesisOpen {
		// Change token type to identifier
		sn.tokens[sn.tokenIndex].Value = sn.tokens[sn.tokenIndex].TokenType.String()
		sn.tokens[sn.tokenIndex].TokenType = lexer.TT_Identifier

		sn.consume()
		sn.analyzeFunctionCall()

		// Not end of expression
		if sn.peek().TokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
		}
		return
	}

	// List
	if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
		sn.analyzeList()

		// Not end of expression
		if sn.peek().TokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
		}
		return
	}

	// Operator missing right side expression
	if sn.peekPrevious().TokenType.IsOperator() {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Operator %s is missing right side expression.", sn.peekPrevious()))
		// Operator missing left side expression
	} else if sn.peek().TokenType.IsBinaryOperator() {
		// Allow only for minus
		if sn.peek().TokenType == lexer.TT_OP_Subtract {
			sn.consume()
			sn.analyzeExpression()
		} else {
			sn.newError(sn.peek(), fmt.Sprintf("Operator %s is missing left side expression.", sn.consume()))

			// Analyze right side expression
			if sn.peek().TokenType.IsLiteral() || sn.peek().TokenType == lexer.TT_Identifier || sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
				sn.analyzeExpression()
				// Right side expression is missing
			} else {
				sn.newError(sn.peekPrevious(), fmt.Sprintf("Operator %s is missing right side expression.", sn.peekPrevious()))
			}
		}
		// Invalid token
	} else {
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in expression.", sn.peek()))

		if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.consume()
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeSubExpression() {
	opening := sn.consume()
	sn.analyzeExpression()

	if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.consume()
		return
	}

	sn.newError(opening, "Missing closing parenthesis of a sub-expression.")
}

func (sn *SyntaxAnalyzer) analyzeList() {
	sn.consume()

	// EOC after opening
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Collect expressions in list
	for sn.peek().TokenType != lexer.TT_DL_BraceClose && sn.peek().TokenType != lexer.TT_EndOfCommand {
		sn.analyzeExpression()

		// Another expression
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()

			// Consume EOC
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.consume()
				// Missing expression
			} else if sn.peek().TokenType == lexer.TT_DL_BraceClose {
				sn.newError(sn.peek(), "Expected expression after comma.")
				break
			}
		}

		// EOC after element
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()

			// Allow only after last element
			if sn.peek().TokenType != lexer.TT_DL_BraceClose {
				sn.newError(sn.peekPrevious(), "There can be EOC (\\n) only after last element. Expected \",\".")
			}
		}
	}

	sn.consume()
}
