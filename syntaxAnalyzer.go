package main

import "fmt"

type SyntaxAnalyzer struct {
	tokens []*Token

	tokenIndex  int
	customTypes map[string]bool

	errorCount uint
}

func NewSyntaxAnalyzer(tokens []*Token) SyntaxAnalyzer {
	return SyntaxAnalyzer{tokens, 0, map[string]bool{}, 0}
}

func (sn *SyntaxAnalyzer) newError(token *Token, message string) {
	sn.errorCount++
	ErrorCodePos(token.position, message)
}

func (sn *SyntaxAnalyzer) peek() *Token {
	return sn.tokens[sn.tokenIndex]
}

func (sn *SyntaxAnalyzer) peekNext() *Token {
	if sn.tokenIndex+1 < len(sn.tokens) {
		return sn.tokens[sn.tokenIndex+1]
	}
	return sn.tokens[sn.tokenIndex]
}

func (sn *SyntaxAnalyzer) peekPrevious() *Token {
	if sn.tokenIndex > 0 {
		return sn.tokens[sn.tokenIndex-1]
	}
	return sn.tokens[0]
}

func (sn *SyntaxAnalyzer) consume() *Token {
	if sn.tokenIndex+1 < len(sn.tokens) {
		sn.tokenIndex++
	}
	return sn.tokens[sn.tokenIndex-1]
}

func (sn *SyntaxAnalyzer) resetTokenPointer() {
	sn.tokenIndex = 0
	sn.consume()
}

func (sn *SyntaxAnalyzer) collectExpression() string {
	i := sn.tokenIndex
	expression := ""

	for i < len(sn.tokens) {
		if sn.tokens[i].tokenType == TT_EndOfCommand || sn.tokens[i].tokenType == TT_EndOfFile {
			return expression
		}

		expression = fmt.Sprintf("%s %s", expression, sn.tokens[i])
		i++
	}

	return expression
}

func (sn *SyntaxAnalyzer) collectLine() string {
	statement := ""

	for sn.peek().tokenType != TT_EndOfFile {
		if sn.peek().tokenType == TT_EndOfCommand && sn.peek().value == "" || sn.peek().tokenType == TT_DL_BraceOpen {
			if len(statement) != 0 {
				return statement[1:]
			}
			return ""
		}
		statement = fmt.Sprintf("%s %s", statement, sn.consume())
	}

	if len(statement) != 0 {
		return statement[1:]
	}
	return ""
}

func (sn *SyntaxAnalyzer) Analyze() {
	// Check StartOfFile
	if sn.peek().tokenType != TT_StartOfFile {
		sn.newError(sn.peek(), "Missing StarOfFile token. This is probably a lexer error.")
	} else {
		sn.consume()
	}

	sn.registerEnumsAndStructs()
	sn.resetTokenPointer()
	sn.analyzeStatementList(false)
}

func (sn *SyntaxAnalyzer) registerEnumsAndStructs() {
	for sn.peek().tokenType != TT_EndOfFile {
		// Register enum
		if sn.peek().tokenType == TT_KW_enum || sn.peek().tokenType == TT_KW_struct {
			sn.consume()

			if sn.peek().tokenType == TT_Identifier {
				sn.customTypes[sn.consume().value] = true
			}
		} else {
			sn.consume()
		}
	}
}

func (sn *SyntaxAnalyzer) lookFor(tokenType TokenType, afterWhat, name string, optional bool) bool {
	if sn.peek().tokenType != tokenType {
		// Skip 1 EOC
		if sn.peek().tokenType == TT_EndOfCommand {
			sn.consume()
		}

		// Check for opening brace after EOCs
		if sn.peek().tokenType == TT_EndOfCommand {
			for sn.peek().tokenType == TT_EndOfCommand {
				sn.consume()
			}

			// Found token
			if sn.peek().tokenType == tokenType {
				sn.newError(sn.peek(), fmt.Sprintf("Too many EOCs (\\n or ;) after %s. Only 0 or 1 EOCs are allowed.", afterWhat))
			} else {
				sn.newError(sn.consume(), fmt.Sprintf("Expected %s after %s.", name, afterWhat))
				return false
			}
		} else if sn.peek().tokenType != tokenType {
			if !optional {
				sn.newError(sn.consume(), fmt.Sprintf("Expected %s after %s.", name, afterWhat))
				return false
			}
		}
	}

	return true
}

func (sn *SyntaxAnalyzer) analyzeStatementList(isScope bool) {
	start := sn.peekPrevious()

	for sn.peek().tokenType != TT_EndOfFile {
		if sn.analyzeStatement(isScope) {
			return
		}
	}

	if isScope {
		sn.newError(start, "Code block is missing closing brace.")
	}
}

func (sn *SyntaxAnalyzer) analyzeStatement(isScope bool) bool {
	switch sn.peek().tokenType {

	case TT_KW_fun: // Function declaration
		sn.analyzeFunctionDeclaration()

	case TT_Identifier: // Identifiers
		sn.analyzeIdentifier()

	case TT_KW_var, TT_KW_bool, TT_KW_int, TT_KW_flt, TT_KW_str: // Variable declarations
		sn.analyzeVariableDeclaration()

	case TT_DL_BraceClose: // Leave scope
		if isScope {
			return true
		}
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\". Expected statement.", sn.consume()))
	
	case TT_DL_ParenthesisClose:
		return true
	
	case TT_DL_BraceOpen: // Enter scope
		sn.analyzeScope()

	case TT_KW_struct: // Struct
		sn.analyzeStructDefinition()

	case TT_KW_enum: // Enum
		sn.analyzeEnumDefinition()

	case TT_KW_if: // If
		sn.analyzeIfStatement()

	case TT_KW_else: // Else
		sn.newError(sn.peek(), "Else statement is missing an if statement.")
		sn.analyzeElseStatement()

	case TT_KW_loop: // Loop
		sn.analyzeLoop()

	case TT_KW_while: // While loop
		sn.analyzeWhileLoop()

	case TT_KW_for: // For loop
		sn.analyzeForLoop()

	case TT_KW_forEach: // ForEach loop
		sn.analyzeForEachLoop()

	case TT_KW_break, TT_KW_drop: // Break, Drop
		sn.consume()

	case TT_EndOfCommand: // Ignore EOCs
		
	default:
		// Collect line and print error
		statement := sn.collectLine()
		sn.newError(sn.peek(), fmt.Sprintf("Invalid statement \"%s\".", statement))
	}

	// Collect tokens after statement
	if sn.peek().tokenType != TT_EndOfCommand && sn.peek().tokenType != TT_EndOfFile {
		if sn.peek().tokenType == TT_DL_ParenthesisClose {
			return true
		}

		statement := sn.collectLine()
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token/s \"%s\" after statement.", statement))
	}
	sn.consume()

	return false
}

func (sn *SyntaxAnalyzer) analyzeForEachLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword forEach, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check type
	if !sn.peek().tokenType.IsVariableType() {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check variable identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check keyword in
	if sn.peek().tokenType != TT_KW_in {
		sn.newError(sn.peek(), fmt.Sprintf("Expected keyword in after variable identifier, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check list identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after variable type, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check closing parenthesis
	if sn.peek().tokenType == TT_DL_ParenthesisClose {
		sn.consume()
	} else {
		for sn.peek().tokenType != TT_DL_ParenthesisClose && sn.peek().tokenType != TT_DL_BraceOpen && sn.peek().tokenType != TT_EndOfCommand {
			sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis, found \"%s\" instead.", sn.consume()))
		}

		// Parenthesis found
		if sn.peek().tokenType == TT_DL_ParenthesisClose {
			sn.consume()
		}
	}

	// Check code block
	if sn.lookFor(TT_DL_BraceOpen, "forEach statement", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeForLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword for, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check condition
	if sn.peek().tokenType == TT_EndOfCommand {
		// Missing condition
		if sn.peek().value == "" {
			sn.newError(sn.peek(), "Expected for loop initialization statement, found \"\\n\" instead.")
			return
		// No condition
		} else {
			sn.consume()
		}
	// Check condition statement
	} else {
		sn.analyzeStatement(false)
	}

	// Check condition
	if sn.peek().tokenType == TT_EndOfCommand {
		// Missing condition
		if sn.peek().value == "" {
			sn.newError(sn.peek(), "For loop missing condition and step statement.")
			return
		} else {
			sn.newError(sn.peek(), "For loop missing condition.")
		}
	// No condition
	} else if sn.peek().tokenType == TT_DL_ParenthesisClose {
		sn.newError(sn.consume(), "For loop missing condition and step statement.")
		return
	// Check condition expression
	} else {
		sn.analyzeExpression()
	}

	// Check step
	if sn.peek().tokenType == TT_EndOfCommand {
		sn.consume()
	// No step
	} else if sn.peek().tokenType == TT_DL_ParenthesisClose {
		return
	}

	sn.analyzeStatement(false)

	// Check closing parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis after condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	if sn.lookFor(TT_DL_BraceOpen, "for loop condition", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeWhileLoop() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword while, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check condition
	if sn.peek().tokenType == TT_EndOfCommand {
		sn.newError(sn.peek(), fmt.Sprintf("Expected condition, found \"%s\" instead.", sn.peek()))
		return
	}
	sn.analyzeExpression()

	// Check closing parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis after condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	if sn.lookFor(TT_DL_BraceOpen, "while loop condition", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeLoop() {
	sn.consume()

	if sn.lookFor(TT_DL_BraceOpen, "keyword loop", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeIfStatement() {
	sn.consume()

	// Check opening parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.newError(sn.peek(), fmt.Sprintf("Expected opening parenthesis after keyword if, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Collect expression
	if sn.peek().tokenType == TT_EndOfCommand || sn.peek().tokenType == TT_EndOfFile {
		sn.newError(sn.peek(), fmt.Sprintf("Expected condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.analyzeExpression()
	}

	// Check closing parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected closing parenthesis after condition, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(TT_DL_BraceOpen, "if statement", "opening brace", false) {
		return
	}

	// Check body
	sn.analyzeScope()

	// Check else statement
	if sn.lookFor(TT_KW_else, "if statement block", "else", true) {
		sn.analyzeElseStatement()
	}
}

func (sn *SyntaxAnalyzer) analyzeElseStatement() {
	sn.consume()

	// Check opening brace
	if sn.lookFor(TT_DL_BraceOpen, "else statement", "opening brace", false) {
		sn.analyzeScope()
	}
}

func (sn *SyntaxAnalyzer) analyzeEnumDefinition() {
	sn.consume()

	// Check identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peek(), "Expected identifier after keyword enum.")

		if sn.peek().tokenType != TT_DL_BraceOpen {
			sn.consume()
		}
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(TT_DL_BraceOpen, "enum identifier", "opening brace", false) {
		return
	}
	sn.consume()

	// Check enums
	for sn.peek().tokenType != TT_EndOfFile {

		// Enum name
		if sn.peek().tokenType == TT_Identifier {
			identifier := sn.consume()

			// Set custom value
			if sn.peek().tokenType == TT_KW_Assign {
				sn.consume()
				sn.analyzeExpression()
			}

			// Allow only EOCs and } after enum name
			if sn.peek().tokenType == TT_EndOfCommand {
				sn.consume()
			} else if sn.peek().tokenType == TT_DL_BraceClose {
				sn.consume()
				break
				// , instead of ;
			} else if sn.peek().tokenType == TT_DL_Comma {
				sn.newError(sn.consume(), "Unexpected token \",\" after enum name. Did you want \";\"?")
				// Invalid token
			} else {
				// Missing =
				if sn.peek().tokenType.IsLiteral() {
					expression := sn.collectExpression()

					sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after enum name. Did you want %s =%s?", sn.peek(), identifier, expression))
					sn.analyzeExpression()
					// Generic error
				} else {
					for sn.peek().tokenType != TT_EndOfFile && sn.peek().tokenType != TT_EndOfCommand && sn.peek().tokenType != TT_DL_BraceClose {
						sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after enum name.", sn.consume()))
					}
				}
			}
			// End of names
		} else if sn.peek().tokenType == TT_DL_BraceClose {
			sn.consume()
			break
			// EOCs
		} else if sn.peek().tokenType == TT_EndOfCommand {
			sn.consume()
			// Invalid token
		} else {
			sn.newError(sn.peek(), "Expected enum name.")
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeStructDefinition() {
	sn.consume()

	// Check identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peek(), "Expected identifier after keyword struct.")

		if sn.peek().tokenType != TT_DL_BraceOpen {
			sn.consume()
		}
	} else {
		sn.consume()
	}

	// Check opening brace
	if !sn.lookFor(TT_DL_BraceOpen, "struct identifier", "opening brace", false) {
		return
	}
	sn.consume()

	// Check properties
	for sn.peek().tokenType != TT_EndOfFile {
		// Property
		if sn.peek().tokenType.IsVariableType() || sn.peek().tokenType == TT_Identifier && sn.customTypes[sn.peek().value] {
			sn.consume()

			// Valid identifier
			if sn.peek().tokenType == TT_Identifier {
				sn.consume()
				// Missing identifier
			} else if sn.peek().tokenType == TT_EndOfCommand {
				sn.newError(sn.peek(), fmt.Sprintf("Expected struct property identifier, found \"%s\" instead.", sn.consume()))
				continue
				// Invalid identifier
			} else {
				sn.newError(sn.peek(), fmt.Sprintf("Expected struct property identifier, found \"%s\" instead.", sn.consume()))
			}

			// , instead of ;
			if sn.peek().tokenType == TT_DL_Comma {
				sn.newError(sn.consume(), "Unexpected token \",\" after enum name. Did you want \";\"?")
				continue
			}

			// Tokens after identifier
			for sn.peek().tokenType != TT_EndOfCommand {
				sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after struct property.", sn.consume()))
			}
			sn.consume()
			// End of properties
		} else if sn.peek().tokenType == TT_DL_BraceClose {
			sn.consume()
			return
			// Empty line
		} else if sn.peek().tokenType == TT_EndOfCommand {
			sn.consume()
			// Invalid token
		} else {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in struct properties.", sn.consume()))
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeVariableDeclaration() {
	sn.consume()

	// Check identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after %s keyword.", sn.peekPrevious()))
	} else {
		sn.consume()
	}

	// Multiple identifiers
	for sn.peek().tokenType == TT_DL_Comma {
		sn.consume()

		// Missing identifier
		if sn.peek().tokenType != TT_Identifier {
			sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after \",\" keyword, found \"%s\" instead.", sn.peek()))

			// Not the end of identifiers
			if sn.peek().tokenType != TT_KW_Assign {
				sn.consume()
				// More identifiers
				if sn.peek().tokenType == TT_DL_Comma {
					continue
				}
			}
			break
		}
		sn.consume()
	}

	// No assign
	if sn.peek().tokenType != TT_KW_Assign {
		if sn.peek().tokenType == TT_EndOfCommand || sn.peek().tokenType == TT_EndOfFile {
			return
		}

		// Collect invalid tokens
		for sn.peek().tokenType != TT_EndOfCommand && sn.peek().tokenType != TT_EndOfFile && sn.peek().tokenType != TT_DL_ParenthesisClose {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after variable declaration.", sn.consume()))
		}
		return
	}

	// Assign
	sn.consume()

	// Missing expression
	if sn.peek().tokenType == TT_EndOfCommand || sn.peek().tokenType == TT_EndOfFile {
		sn.newError(sn.peek(), "Assign statement is missing assigned expression.")
		return
	}

	sn.analyzeExpression()

	// Collect invalid tokens
	for sn.peek().tokenType != TT_EndOfCommand && sn.peek().tokenType != TT_EndOfFile && sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after variable declaration.", sn.consume()))
	}
}

func (sn *SyntaxAnalyzer) analyzeFunctionDeclaration() {
	sn.consume()

	// Collect identifier
	if sn.peek().tokenType != TT_Identifier {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected function identifier after fun keyword, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	// Check for opening parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected opening parenthesis after function identifier, found \"%s\" instead.", sn.peek()))
	} else {
		sn.consume()
	}

	if sn.peek().tokenType != TT_DL_ParenthesisOpen {
		sn.analyzeParameters()
	}

	// Check for closing parenthesis
	var closingParent *Token = nil
	if sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Expected closing parenthesis after function identifier, found \"%s\" instead.", sn.peek()))
	} else {
		closingParent = sn.consume()
	}

	// Check return type
	if sn.peek().tokenType == TT_KW_returns {
		sn.consume()

		// Missing return type
		if sn.peek().tokenType == TT_EndOfCommand || sn.peek().tokenType == TT_DL_BraceOpen {
			sn.newError(sn.peek(), fmt.Sprintf("Expected return type after keyword ->, found \"%s\" instead.", sn.peek()))
		} else {
			// Check if type is valid
			if !sn.peek().tokenType.IsVariableType() && !(sn.peek().tokenType == TT_Identifier && sn.customTypes[sn.peek().value]) {
				sn.newError(sn.peek(), fmt.Sprintf("Expected return type after keyword ->, found \"%s\" instead.", sn.peek()))
			}
			sn.consume()
		}
	}

	// Check for start of scope
	if sn.peek().tokenType == TT_DL_BraceOpen {
		sn.analyzeScope()
		return
	}

	if sn.peek().tokenType == TT_EndOfCommand {
		sn.consume()
		if sn.peek().tokenType == TT_DL_BraceOpen {
			sn.analyzeScope()
			return
		}
	}

	// Scope not found
	// Multiple EOCs
	if sn.peek().tokenType == TT_EndOfCommand {
		for sn.peek().tokenType == TT_EndOfCommand {
			sn.consume()
		}

		if sn.peek().tokenType == TT_DL_BraceOpen {
			sn.newError(closingParent, "Too many EOCs (\\n or ;) after function header. Only 0 or 1 EOCs are allowed.")
			sn.analyzeScope()
			return
		}
	}

	// Invalid tokens
	sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" after function header. Expected code block.", sn.peek()))
}

func (sn *SyntaxAnalyzer) analyzeIdentifier() {
	// Enum variable declaration
	if sn.customTypes[sn.peek().value] {
		sn.analyzeVariableDeclaration()
		return
	}

	sn.consume()

	// Function call
	if sn.peek().tokenType == TT_DL_ParenthesisOpen {
		sn.analyzeFunctionCall()
		// Assignment
	} else if sn.peek().tokenType.IsAssignKeyword() {
		sn.analyzeAssignment()
	}
}

func (sn *SyntaxAnalyzer) analyzeAssignment() {
	sn.consume()

	sn.analyzeExpression()
}

func (sn *SyntaxAnalyzer) analyzeFunctionCall() {
	sn.consume()
	sn.analyzeAruments()

	// Check closing parenthesis
	if sn.peek().tokenType != TT_DL_ParenthesisClose {
		sn.newError(sn.peek(), fmt.Sprintf("Expected \")\" after function call arguments, found \"%s\" instead.", sn.peek()))

		for sn.peek().tokenType != TT_EndOfCommand {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in function call.", sn.consume()))
		}

		return
	}
	sn.consume()
}

func (sn *SyntaxAnalyzer) analyzeAruments() {
	for sn.peek().tokenType != TT_EndOfFile {
		if sn.peek().tokenType == TT_DL_ParenthesisClose || sn.peek().tokenType == TT_EndOfCommand {
			return
		}

		sn.analyzeExpression()

		// Next argument
		if sn.peek().tokenType == TT_DL_Comma {
			sn.consume()
			// End of arguments
		} else if sn.peek().tokenType == TT_DL_ParenthesisClose || sn.peek().tokenType == TT_EndOfCommand {
			return
			// Invalid token
		} else {
			sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in argument.", sn.consume()))
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeParameters() {
	for sn.peek().tokenType != TT_EndOfFile && sn.peek().tokenType != TT_EndOfCommand {
		// Check type
		if !sn.peek().tokenType.IsVariableType() {
			sn.newError(sn.peek(), fmt.Sprintf("Expected variable type at start of parameter, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// Check identifier
		if sn.peek().tokenType != TT_Identifier {
			sn.newError(sn.peek(), fmt.Sprintf("Expected parameter identifier after parameter type, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// No default value
		if sn.peek().tokenType == TT_DL_Comma {
			sn.consume()
			continue
		} else if sn.peek().tokenType == TT_DL_ParenthesisClose {
			return
		}

		// Check assign
		if sn.peek().tokenType != TT_KW_Assign {
			sn.newError(sn.peek(), fmt.Sprintf("Expected \"=\" or \",\" after parameter identifier, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// Check default value
		if !sn.peek().tokenType.IsLiteral() {
			sn.newError(sn.peek(), fmt.Sprintf("Expected default value literal after = in function parameter, found \"%s\" instead.", sn.peek()))
		} else {
			sn.consume()
		}

		// Next parameter
		if sn.peek().tokenType == TT_DL_Comma {
			sn.consume()
			continue
		} else if sn.peek().tokenType == TT_DL_ParenthesisClose {
			return
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeScope() {
	sn.consume()
	sn.analyzeStatementList(true)
	sn.consume()
}

func (sn *SyntaxAnalyzer) analyzeExpression() {
	// No expression
	if sn.peek().tokenType == TT_EndOfCommand {
		sn.newError(sn.peek(), "Expected expression, found EOC instead.")
		return
	}

	// Operator
	if sn.peek().tokenType.IsUnaryOperator() {
		sn.consume()
		sn.analyzeExpression()
		return
	}
	// Literal or identifier
	if sn.peek().tokenType.IsLiteral() || sn.peek().tokenType == TT_Identifier {
		sn.consume()

		if sn.peek().tokenType.IsBinaryOperator() {
			sn.consume()
			sn.analyzeExpression()
		}
		return
	}
	// Sub-Expression
	if sn.peek().tokenType == TT_DL_ParenthesisOpen {
		sn.analyzeSubExpression()

		if sn.peek().tokenType.IsBinaryOperator() {
			sn.analyzeExpression()
		}
		return
	}

	// Operator missing right side expression
	if sn.peekPrevious().tokenType.IsOperator() {
		sn.newError(sn.peekPrevious(), fmt.Sprintf("Operator %s is missing right side expression.", sn.peekPrevious()))
		// Operator missing left side expression
	} else if sn.peek().tokenType.IsBinaryOperator() {
		// Allow only for minus
		if sn.peek().tokenType == TT_OP_Subtract {
			sn.consume()
			sn.analyzeExpression()
		} else {
			sn.newError(sn.peek(), fmt.Sprintf("Operator %s is missing left side expression.", sn.consume()))

			// Analyze right side expression
			if sn.peek().tokenType.IsLiteral() || sn.peek().tokenType == TT_Identifier || sn.peek().tokenType == TT_DL_ParenthesisOpen {
				sn.analyzeExpression()
				// Right side expression is missing
			} else {
				sn.newError(sn.peekPrevious(), fmt.Sprintf("Operator %s is missing right side expression.", sn.peekPrevious()))
			}
		}
		// Invalid token
	} else {
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\" in expression.", sn.peek()))

		if sn.peek().tokenType != TT_DL_ParenthesisClose {
			sn.consume()
		}
	}
}

func (sn *SyntaxAnalyzer) analyzeSubExpression() {
	opening := sn.consume()
	sn.analyzeExpression()

	if sn.peek().tokenType == TT_DL_ParenthesisClose {
		sn.consume()
		return
	}

	sn.newError(opening, "Missing closing parenthesis of a sub-expression.")
}
