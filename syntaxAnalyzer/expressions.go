package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeRestOfExpression() {
	if sn.peek().TokenType == lexer.TT_OP_QuestionMark || sn.peek().TokenType == lexer.TT_OP_Not {
		sn.consume()
	}

	if sn.peek().TokenType.IsBinaryOperator() {
		sn.consume()
		sn.analyzeExpression()
	}
}

func (sn *SyntaxAnalyzer) analyzeExpression() {
	// No expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Unary operator
	if sn.peek().TokenType.IsUnaryOperator() {
		sn.consume()
		sn.analyzeExpression()
		return
	}

	// Literal
	if sn.peek().TokenType.IsLiteral() {
		sn.consume()
		sn.analyzeRestOfExpression()
		return
	}

	// Set
	if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
		sn.analyzeSet()
		sn.analyzeRestOfExpression()
		return
	}

	// List
	if sn.peek().TokenType == lexer.TT_DL_BracketOpen {
		sn.analyzeList()
		sn.analyzeRestOfExpression()
		return
	}

	// Object literal
	if sn.peek().TokenType == lexer.TT_Identifier && sn.peekNext().TokenType == lexer.TT_DL_BraceOpen {
		sn.analyzeObject()
		return
	}

	// Set/List with it's type specified
	if sn.peek().TokenType.IsCompositeType() {
		sn.analyzeCompositeType()

		if sn.peek().TokenType == lexer.TT_DL_BracketOpen {
			sn.analyzeList()
		} else if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.analyzeSet()
		} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.newError(sn.peek(), "Missing literal after type hint.")
		} else {
			sn.analyzeExpression()
		}

		sn.analyzeRestOfExpression()
		return
	}

	// Identifier
	if sn.peek().TokenType == lexer.TT_Identifier {
		sn.analyzeIdentifier()
		sn.analyzeRestOfExpression()
		return
	}

	// Sub-Expression
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeSubExpression()
		sn.analyzeRestOfExpression()
		return
	}

	// Function call of function with same name as keyword
	if (sn.peek().TokenType == lexer.TT_KW_int || sn.peek().TokenType == lexer.TT_KW_flt || sn.peek().TokenType == lexer.TT_KW_str) && sn.peekNext().TokenType == lexer.TT_DL_ParenthesisOpen {
		// Change token type to identifier
		sn.tokens[sn.tokenIndex].Value = sn.tokens[sn.tokenIndex].TokenType.String()
		sn.tokens[sn.tokenIndex].TokenType = lexer.TT_Identifier

		sn.consume()
		sn.analyzeFunctionCall()

		sn.analyzeRestOfExpression()
		return
	}

	// Operator missing right side expression
	if sn.peekPrevious().TokenType.IsOperator() {
		sn.newError(sn.peekPrevious(), "Operator "+sn.peekPrevious().String()+" is missing right side expression.")
		// Operator missing left side expression
	} else if sn.peek().TokenType.IsBinaryOperator() {
		// Allow only for minus
		if sn.peek().TokenType == lexer.TT_OP_Subtract {
			sn.consume()
			sn.analyzeExpression()
		} else {
			sn.newError(sn.peek(), "Operator "+sn.consume().String()+" is missing left side expression.")

			// Analyze right side expression
			if sn.peek().TokenType.IsLiteral() || sn.peek().TokenType == lexer.TT_Identifier || sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
				sn.analyzeExpression()
				// Right side expression is missing
			} else {
				sn.newError(sn.peekPrevious(), "Operator "+sn.peekPrevious().String()+" is missing right side expression.")
			}
		}
		// Invalid token
	} else {
		sn.newError(sn.peek(), "Unexpected token \""+sn.peek().String()+"\" in expression.")

		if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.consume()
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeSubExpression() {
	opening := sn.consume() // (
	sn.analyzeExpression()

	if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
		sn.consume() // )
		return
	}

	sn.newError(opening, "Missing closing parenthesis of a sub-expression.")
}

func (sn *SyntaxAnalyzer) analyzeList() {
	sn.consume() // [

	// EOC after opening
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Collect expressions in list
	for sn.peek().TokenType != lexer.TT_DL_BracketClose && sn.peek().TokenType != lexer.TT_EndOfCommand {
		sn.analyzeExpression()

		// Another expression
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()

			// Consume EOC
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.consume()
				// Missing expression
			} else if sn.peek().TokenType == lexer.TT_DL_BracketClose {
				sn.newError(sn.peek(), "Expected expression after comma.")
				break
			}
		}

		// EOC after element
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()

			// Allow only after last element
			if sn.peek().TokenType != lexer.TT_DL_BracketClose {
				sn.newError(sn.peekPrevious(), "There can be EOC (\\n) only after last element. Expected \",\".")
			}
		}
	}

	sn.consume() // ]
}

func (sn *SyntaxAnalyzer) analyzeSet() {
	sn.consume() // {
	sn.consumeEOCs()

	for sn.peek().TokenType != lexer.TT_DL_BraceClose {
		// Analyze expression
		sn.analyzeExpression()

		// Comma after element
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()

			// Closing brace right after comma
			if sn.peek().TokenType == lexer.TT_DL_BraceClose {
				sn.newError(sn.peek(), "Expected expression or EOC after comma.")
				break
			}

			sn.consumeEOCs()

			// No more elements after comma
			if sn.peek().TokenType == lexer.TT_DL_BraceClose {
				sn.consume()
				break
			}

			continue
		}

		sn.consumeEOCs()

		// No comma after element and no closing brace
		if sn.peek().TokenType != lexer.TT_DL_BraceClose {
			sn.newError(sn.peek(), "Expected closing brace (\"}\") after last set element.")
		}
	}

	sn.consume() // }
}

func (sn *SyntaxAnalyzer) analyzeObject() {
	sn.consume() // ID
	sn.consume() // {

	// Consume EOCs
	for sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}

	// Collect properties
	for sn.peek().TokenType != lexer.TT_EndOfCommand {
		// End of properties
		if sn.peek().TokenType == lexer.TT_DL_BraceClose {
			sn.consume()
			break
		}

		// Consume property label
		if sn.peek().TokenType == lexer.TT_Identifier && sn.peekNext().TokenType == lexer.TT_DL_Colon {
			sn.consume() // ID
			sn.consume() // ,
		}

		// Collect property
		sn.analyzeExpression()

		// End of properties
		if sn.peek().TokenType == lexer.TT_DL_BraceClose {
			sn.consume()
			break
			// More properties
		} else if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()
		}

		// Consume EOCs
		for sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}
	}
}
