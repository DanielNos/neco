package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) analyzeListType() {
	// Collect list
	sn.consume()

	// Collect <
	if sn.peek().TokenType == lexer.TT_OP_Lower {
		sn.consume()
	} else {
		sn.newError(sn.peek(), "Expected opening < after list type.")
	}

	// Collect sub-type
	if sn.peek().TokenType.IsVariableType() {
		if sn.peek().TokenType == lexer.TT_KW_list {
			sn.analyzeListType()
		} else {
			sn.consume()
		}
	} else {
		sn.newError(sn.peek(), "Expected subtype in composite data type.")
	}

	// Collect >
	if sn.peek().TokenType == lexer.TT_OP_Greater {
		sn.consume()
	} else {
		sn.newError(sn.peek(), "Expected closing > aftrer data type.")
	}
}

func (sn *SyntaxAnalyzer) analyzeVariableDeclaration(constant bool) {
	// Collect type
	if sn.peek().TokenType == lexer.TT_KW_list {
		sn.analyzeListType()
	} else {
		sn.consume()
	}

	// Check identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		if constant {
			sn.newError(sn.peekPrevious(), "Expected variable identifier after const keyword.")
		}
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after %s keyword.", sn.peekPrevious()))
	} else {
		sn.consume()
	}

	// Multiple identifiers
	for sn.peek().TokenType == lexer.TT_DL_Comma {
		sn.consume()

		// Missing identifier
		if sn.peek().TokenType != lexer.TT_Identifier {
			sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after \",\" keyword, found \"%s\" instead.", sn.peek()))

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
		if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_EndOfFile {
			return
		}

		// Collect invalid tokens
		for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after variable declaration.", sn.consume()))
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
	for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after variable declaration.", sn.consume()))
	}
}

func (sn *SyntaxAnalyzer) analyzeFunctionDeclaration() {
	sn.consume()

	// Collect identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected function identifier after fun keyword, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check for opening parenthesis
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisOpen {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected opening parenthesis after function identifier, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.analyzeParameters()
	}

	// Check for closing parenthesis
	var closingParent *lexer.Token = nil
	if sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected closing parenthesis after function identifier, found \"%s\" instead.", sn.peek()))
	} else {
		closingParent = sn.consume()
	}

	// Check return type
	if sn.peek().TokenType == lexer.TT_KW_returns {
		sn.consume()

		// Missing return type
		if sn.peek().TokenType == lexer.TT_EndOfCommand || sn.peek().TokenType == lexer.TT_DL_BraceOpen {
			sn.newError(sn.peek(), fmt.Sprintf("Expected return type after keyword ->, found \"%s\" instead.", sn.peek()))
		} else {
			// Check if type is valid
			if !sn.peek().TokenType.IsVariableType() && !(sn.peek().TokenType == lexer.TT_Identifier && sn.customTypes[sn.peek().Value]) {
				sn.newError(sn.peek(), fmt.Sprintf("Expected return type after keyword ->, found \"%s\" instead.", sn.peek()))
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
	sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after function header. Expected code block.", sn.peek()))
}

func (sn *SyntaxAnalyzer) analyzeParameters() {
	for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_EndOfCommand {
		// Check type
		if !sn.peek().TokenType.IsVariableType() {
			sn.newError(sn.peek(), fmt.Sprintf("Expected variable type at start of parameters, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// Check identifier
		if sn.peek().TokenType != lexer.TT_Identifier {
			sn.newError(sn.peek(), fmt.Sprintf("Expected parameter identifier after parameter type, found \"%s\" instead.", sn.peek()))
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
			sn.newError(sn.peek(), fmt.Sprintf("Expected \"=\" or \",\" after parameter identifier, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// Check default value
		if !sn.peek().TokenType.IsLiteral() {
			sn.newError(sn.peek(), fmt.Sprintf("Expected default value literal after = in function parameter, found \"%s\" instead.", sn.peek()))
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
