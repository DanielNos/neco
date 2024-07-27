package syntaxAnalyzer

import (
	"github.com/DanielNos/neco/lexer"
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
		sn.consumeEOCs()

		// Property
		if sn.peek().TokenType.IsVariableType() || sn.peek().TokenType == lexer.TT_Identifier && sn.customTypes[sn.peek().Value] {
			sn.consume()

			// Valid identifier
			if sn.peek().TokenType == lexer.TT_Identifier {
				sn.consume()
				// Missing identifier
			} else if sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), "Expected struct property identifier, found \""+sn.consume().String()+"\" instead.")
				continue
				// Invalid identifier
			} else {
				sn.newError(sn.peek(), "Expected struct property identifier, found \""+sn.consume().String()+"\" instead.")
			}

			// More identifiers of same type
			for sn.peek().TokenType == lexer.TT_DL_Comma {
				sn.consume()

				// No identifier
				if sn.peek().TokenType != lexer.TT_Identifier {
					sn.newError(sn.peek(), "Expected property identifier after comma.")

					// End of property
					if sn.peek().TokenType == lexer.TT_EndOfCommand {
						continue
					}
					// Consume identifier
				} else {
					sn.consume()
				}
			}

			// Tokens after identifier
			for sn.peek().TokenType != lexer.TT_EndOfCommand {
				sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" after struct property.")
			}
			sn.consume()
			// End of properties
		} else if sn.peek().TokenType == lexer.TT_DL_BraceClose {
			sn.consume()
			return
			// Invalid token
		} else {
			sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" in struct properties.")
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

					sn.newError(sn.peek(), "Unexpected token \""+sn.peek().String()+"\" after enum name. Did you want "+identifier.String()+" = "+expression+"?")
					sn.analyzeExpression()
					// Generic error
				} else {
					for sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_DL_BraceClose {
						sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\" after enum name.")
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
