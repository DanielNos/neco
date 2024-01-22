package syntaxAnalyzer

import (
	"fmt"
	"neko/lexer"
)

func (sn *SyntaxAnalyzer) analyzeExpression() {
	// No expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.newError(sn.peek(), "Expected expression, found EOC instead.")
		return
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
