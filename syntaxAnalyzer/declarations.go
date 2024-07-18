package syntaxAnalyzer

import (
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeVariableDeclaration(constant bool) {
	// Analyze type
	sn.analyzeType()

	// Check identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		if constant {
			sn.newError(sn.peekPrevious(), "Expected variable identifier after const keyword.")
		}
		sn.newError(sn.peek(), "Expected variable identifier after "+sn.peekPrevious().String()+" keyword.")
	} else {
		sn.consume()
	}

	// Multiple identifiers
	for sn.peek().TokenType == lexer.TT_DL_Comma {
		sn.consume()

		// Missing identifier
		if sn.peek().TokenType != lexer.TT_Identifier {
			sn.newError(sn.peek(), "Expected variable identifier after \",\" keyword, found \""+sn.peek().String()+"\" instead.")

			// Not the end of identifiers
			if sn.peek().TokenType != lexer.TT_KW_Assign {
				sn.consume()
				// More identifiers
				if sn.peek().TokenType == lexer.TT_DL_Comma {
					continue
				}
			}
			break
		}
		sn.consume()
	}

	// No assign
	if sn.peek().TokenType != lexer.TT_KW_Assign {
		if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_DL_BraceClose || sn.peek().TokenType == lexer.TT_EndOfFile {
			return
		}

		// Collect invalid tokens
		for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" after variable declaration.")
		}
		return
	}

	// Assign
	sn.consume()

	// Missing expression
	if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_EndOfFile {
		sn.newError(sn.peek(), "Assign statement is missing assigned expression.")
		return
	}

	sn.analyzeExpression()

	// Collect invalid tokens
	for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_DL_BraceClose && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" after variable declaration.")
	}
}

func (sn *SyntaxAnalyzer) analyzeFunctionDeclaration() {
	sn.consume()

	// Collect identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peekPrevious(), "Expected function identifier after fun keyword, found \""+sn.peek().String()+"\" instead.")
	} else {
		sn.consume()
	}

	// Check for opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peekPrevious(), "Expected opening parenthesis after function identifier, found \""+sn.peek().String()+"\" instead.")
	} else {
		sn.consume()
	}

	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.analyzeParameters()
	}

	// Check for closing parenthesis
	var closingParent *lexer.Token = nil
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peekPrevious(), "Expected closing parenthesis after function identifier, found \""+sn.peek().String()+"\" instead.")
	} else {
		closingParent = sn.consume()
	}

	// Check return type
	if sn.peek().TokenType == lexer.TT_KW_returns {
		sn.consume()

		// Missing return type
		if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.newError(sn.peek(), "Expected return type after keyword ->, found \""+sn.peek().String()+"\" instead.")
		} else {
			// Check if type is valid
			if !sn.peek().TokenType.IsVariableType() && !(sn.peek().TokenType == lexer.TT_Identifier && sn.customTypes[sn.peek().Value]) {
				sn.newError(sn.peek(), "Expected return type after keyword ->, found \""+sn.peek().String()+"\" instead.")
			}
			sn.consume()
		}
	}

	// Check for start of scope
	if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
		sn.analyzeScope()
		return
	}

	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
		if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.analyzeScope()
			return
		}
	}

	// Scope not found
	// Multiple EOCs
	if sn.peek().TokenType == lexer.TT_EndOfCommand {
		for sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}

		if sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.newError(closingParent, "Too many EOCs (\\n or ;) after function header. Only 0 or 1 EOCs are allowed.")
			sn.analyzeScope()
			return
		}
	}

	// Invalid tokens
	sn.newError(sn.peek(), "Unexpected token \""+sn.peek().String()+"\" after function header. Expected code block.")
}

func (sn *SyntaxAnalyzer) analyzeParameters() {
	for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_EndOfCommand {
		// Check type
		if !sn.peek().TokenType.IsVariableType() {
			sn.newError(sn.peek(), "Expected variable type at start of parameters, found \""+sn.peek().String()+"\" instead.")
		} else {
			sn.consume()
		}

		// Check identifier
		if sn.peek().TokenType != lexer.TT_Identifier {
			sn.newError(sn.peek(), "Expected parameter identifier after parameter type, found \""+sn.peek().String()+"\" instead.")
		} else {
			sn.consume()
		}

		// No default value
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()

			// Multiple identifiers
			for sn.peek().TokenType == lexer.TT_Identifier {
				sn.consume()

				if sn.peek().TokenType == lexer.TT_DL_Comma {
					sn.consume()
				}
			}
			continue
		} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			return
		}

		// Check assign
		if sn.peek().TokenType != lexer.TT_KW_Assign {
			sn.newError(sn.peek(), "Expected \"=\" or \",\" after parameter identifier, found \""+sn.peek().String()+"\" instead.")
		} else {
			sn.consume()
		}

		// Check default value
		if !sn.peek().TokenType.IsLiteral() {
			sn.newError(sn.peek(), "Expected default value literal after = in function parameter, found \""+sn.peek().String()+"\" instead.")
		} else {
			sn.consume()
		}

		// Next parameter
		if sn.peek().TokenType == lexer.TT_DL_Comma {
			sn.consume()
			continue
		} else if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			return
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeType() {
	if sn.peek().TokenType.IsCompositeType() {
		sn.analyzeCompositeType()
	} else {
		sn.consume()
	}

	if sn.peek().TokenType == lexer.TT_OP_QuestionMark {
		sn.consume()
	}
}

func (sn *SyntaxAnalyzer) analyzeCompositeType() {
	// Consume type
	sn.consume()

	// Consume opening token
	if sn.peek().TokenType == lexer.TT_OP_Lower {
		sn.consume()
	} else {
		sn.newError(sn.peek(), "Expected \"<\" after composite data type.")
	}

	// Analyze sub-type
	if sn.peek().TokenType.IsVariableType() {
		sn.analyzeType()
	} else {
		sn.newError(sn.peek(), "Expected subtype in composite data type.")
	}

	// Consume closing token
	if sn.peek().TokenType == lexer.TT_OP_Greater {
		sn.consume()
	} else {
		sn.newError(sn.peek(), "Expected \">\" after composite data type sub-type.")
	}
}
