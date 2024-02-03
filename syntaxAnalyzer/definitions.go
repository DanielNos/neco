package syntaxAnalyzer

import (
	"fmt"
	"neco/lexer"
)

func (sn *SyntaxAnalyzer) registerEnumsAndStructs() {
	for sn.peek().TokenType != lexer.TT_EndOfFile {
		// Register enum
		if sn.peek().TokenType == lexer.TT_KW_enum || sn.peek().TokenType == lexer.TT_KW_struct {
			sn.consume()

			if sn.peek().TokenType == lexer.TT_Identifier {
				sn.customTypes[sn.consume().Value] = true
			}
		} else {
			sn.consume()
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeStructDefinition() {
	sn.consume()

	// Check identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peek(), "Expected identifier after keyword struct.")

		if sn.peek().TokenType != lexer.TT_DL_BraceOpen {
			sn.consume()
		}
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(lexer.TT_DL_BraceOpen, "struct identifier", "opening brace", false) {
		return
	}
	sn.consume()

	// Check properties
	for sn.peek().TokenType != lexer.TT_EndOfFile {
		// Property
		if sn.peek().TokenType.IsVariableType() || sn.peek().TokenType == lexer.TT_Identifier && sn.customTypes[sn.peek().Value] {
			sn.consume()

			// Valid identifier
			if sn.peek().TokenType == lexer.TT_Identifier {
				sn.consume()
				// Missing identifier
			} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), fmt.Sprintf("Expected struct property identifier, found \"%s\" instead.", sn.consume()))
				continue
				// Invalid identifier
			} else {
				sn.newError(sn.peek(), fmt.Sprintf("Expected struct property identifier, found \"%s\" instead.", sn.consume()))
			}

			// , instead of ;
			if sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.newError(sn.consume(), "Unexpected token \",\" after enum name. Did you want \";\"?")
				continue
			}

			// Tokens after identifier
			for sn.peek().TokenType != lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after struct property.", sn.consume()))
			}
			sn.consume()
			// End of properties
		} else if sn.peek().TokenType == lexer.TT_DL_BraceClose {
			sn.consume()
			return
			// Empty line
		} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
			// Invalid token
		} else {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in struct properties.", sn.consume()))
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeEnumDefinition() {
	sn.consume()

	// Check identifier
	if sn.peek().TokenType != lexer.TT_Identifier {
		sn.newError(sn.peek(), "Expected identifier after keyword enum.")

		if sn.peek().TokenType != lexer.TT_DL_BraceOpen {
			sn.consume()
		}
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(lexer.TT_DL_BraceOpen, "enum identifier", "opening brace", false) {
		return
	}
	sn.consume()

	// Check enums
	for sn.peek().TokenType != lexer.TT_EndOfFile {

		// Enum name
		if sn.peek().TokenType == lexer.TT_Identifier {
			identifier := sn.consume()

			// Set custom value
			if sn.peek().TokenType == lexer.TT_KW_Assign {
				sn.consume()
				sn.analyzeExpression()
			}

			// Allow only EOCs and } after enum name
			if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.consume()
			} else if sn.peek().TokenType == lexer.TT_DL_BraceClose {
				sn.consume()
				break
				// , instead of ;
			} else if sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.newError(sn.consume(), "Unexpected token \",\" after enum name. Did you want \";\"?")
				// Invalid token
			} else {
				// Missing =
				if sn.peek().TokenType.IsLiteral() {
					expression := sn.collectExpression()

					sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after enum name. Did you want %s = %s?", sn.peek(), identifier, expression))
					sn.analyzeExpression()
					// Generic error
				} else {
					for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_DL_BraceClose {
						sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after enum name.", sn.consume()))
					}
				}
			}
			// End of names
		} else if sn.peek().TokenType == lexer.TT_DL_BraceClose {
			sn.consume()
			break
			// EOCs
		} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
			// Invalid token
		} else {
			sn.newError(sn.peek(), "Expected enum name.")
		}
	}
}