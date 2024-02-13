package syntaxAnalyzer

import (
	"fmt"
	"os"

	"neco/errors"
	"neco/lexer"
	"neco/logger"
)

type SyntaxAnalyzer struct {
	tokens []*lexer.Token

	tokenIndex  int
	customTypes map[string]bool

	ErrorCount      uint
	totalErrorCount uint
}

func NewSyntaxAnalyzer(tokens []*lexer.Token, previousErrors uint) SyntaxAnalyzer {
	return SyntaxAnalyzer{tokens, 0, map[string]bool{}, 0, previousErrors}
}

func (sn *SyntaxAnalyzer) newError(token *lexer.Token, message string) {
	if sn.ErrorCount == 0 || sn.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	sn.ErrorCount++
	logger.ErrorCodePos(token.Position, message)

	// Too many errors
	if sn.ErrorCount+sn.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Syntax analysis has aborted due to too many errors. It has failed with %d errors.", sn.ErrorCount))
	}
}

func (sn *SyntaxAnalyzer) newErrorFromTo(line, startChar, endChar uint, message string) {
	if sn.ErrorCount == 0 || sn.totalErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	sn.ErrorCount++
	logger.ErrorPos(sn.peek().Position.File, line, startChar, endChar, message)

	// Too many errors
	if sn.ErrorCount+sn.totalErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Syntax analysis has aborted due to too many errors. It has failed with %d errors.", sn.ErrorCount))
	}
}

func (sn *SyntaxAnalyzer) peek() *lexer.Token {
	return sn.tokens[sn.tokenIndex]
}

func (sn *SyntaxAnalyzer) peekNext() *lexer.Token {
	if sn.tokenIndex+1 < len(sn.tokens) {
		return sn.tokens[sn.tokenIndex+1]
	}
	return sn.tokens[sn.tokenIndex]
}

func (sn *SyntaxAnalyzer) peekPrevious() *lexer.Token {
	if sn.tokenIndex > 0 {
		return sn.tokens[sn.tokenIndex-1]
	}
	return sn.tokens[0]
}

func (sn *SyntaxAnalyzer) consume() *lexer.Token {
	if sn.tokenIndex+1 < len(sn.tokens) {
		sn.tokenIndex++
	}
	return sn.tokens[sn.tokenIndex-1]
}

func (sn *SyntaxAnalyzer) rewind() {
	if sn.tokenIndex > 0 {
		sn.tokenIndex--
	}
}

func (sn *SyntaxAnalyzer) resetTokenPointer() {
	sn.tokenIndex = 0
	sn.consume()
}

func (sn *SyntaxAnalyzer) collectExpression() string {
	i := sn.tokenIndex
	expression := ""

	for i < len(sn.tokens) {
		if sn.tokens[i].TokenType == lexer.TT_EndOfCommand || sn.tokens[i].TokenType == lexer.TT_EndOfFile {
			// Return expression string
			if len(expression) > 0 {
				return expression[1:]
			} else {
				return ""
			}
		}

		expression = fmt.Sprintf("%s %s", expression, sn.tokens[i])
		i++
	}

	// Return expression string
	if len(expression) > 0 {
		return expression[1:]
	} else {
		return ""
	}
}

func (sn *SyntaxAnalyzer) collectLine() string {
	statement := ""

	for sn.peek().TokenType != lexer.TT_EndOfFile {
		if sn.peek().TokenType == lexer.TT_EndOfCommand && sn.peek().Value == "" || sn.peek().TokenType == lexer.TT_DL_BraceOpen {
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
	if sn.peek().TokenType != lexer.TT_StartOfFile {
		sn.newError(sn.peek(), "Missing StarOfFile token. This is probably a lexer error.")
	} else {
		sn.consume()
	}

	sn.registerEnumsAndStructs()
	sn.resetTokenPointer()
	sn.analyzeStatementList(false)
}

func (sn *SyntaxAnalyzer) lookFor(tokenType lexer.TokenType, afterWhat, name string, optional bool) bool {
	if sn.peek().TokenType != tokenType {
		// Skip 1 EOC
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			sn.consume()
		}

		// Check for opening brace after EOCs
		if sn.peek().TokenType == lexer.TT_EndOfCommand {
			for sn.peek().TokenType == lexer.TT_EndOfCommand {
				sn.consume()
			}

			// Found token
			if sn.peek().TokenType == tokenType {
				sn.newError(sn.peek(), fmt.Sprintf("Too many EOCs (\\n or ;) after %s. Only 0 or 1 EOCs are allowed.", afterWhat))
			} else {
				sn.newError(sn.peek(), fmt.Sprintf("Expected %s after %s.", name, afterWhat))
				sn.rewind()
				return false
			}
		} else if sn.peek().TokenType != tokenType {
			if !optional {
				sn.newError(sn.peek(), fmt.Sprintf("Expected %s after %s.", name, afterWhat))
				sn.rewind()
				return false
			}
		}
	}

	return true
}

func (sn *SyntaxAnalyzer) analyzeStatement(isScope bool) bool {
	switch sn.peek().TokenType {

	case lexer.TT_KW_fun: // Function declaration
		sn.analyzeFunctionDeclaration()

	case lexer.TT_Identifier: // Identifiers
		// Function call
		if sn.peekNext().TokenType == lexer.TT_DL_ParenthesisOpen {
			sn.analyzeIdentifier()
			// Variable
		} else {
			// Assignment
			if sn.peekNext().TokenType.IsAssignKeyword() {
				sn.consume()
				sn.analyzeAssignment()
				break
			}
			// Assignment to multiple variables
			if sn.peekNext().TokenType == lexer.TT_DL_Comma {
				sn.consume()
				for sn.peek().TokenType == lexer.TT_DL_Comma {
					sn.consume()
					if sn.peek().TokenType != lexer.TT_Identifier {
						sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after comma, found %s instead.", sn.peek()))
						break
					}
					sn.consume()
				}

				// No = after variables
				if !sn.peek().TokenType.IsAssignKeyword() {
					sn.newError(sn.peek(), fmt.Sprintf("Expected = after list of variable identifiers, found %s instead.", sn.peek()))
					break
				}

				// Assignment
				sn.analyzeAssignment()
				break
			}

			// Declare custom variable
			if sn.customTypes[sn.peek().Value] {
				sn.consume()

				// Check identifier
				if sn.peek().TokenType != lexer.TT_Identifier {
					sn.newError(sn.peek(), fmt.Sprintf("Expected variable identifier after type %s, found \"%s\" instead.", sn.peekPrevious().Value, sn.peek()))
				} else {
					sn.consume()
				}

				// Assign to it
				if sn.peek().TokenType == lexer.TT_KW_Assign {
					sn.analyzeAssignment()
				}
				break
			}

			// Expression
			startChar := sn.peek().Position.StartChar
			sn.analyzeExpression()
			sn.newErrorFromTo(sn.peek().Position.Line, startChar, sn.peek().Position.StartChar, "Expression can't be a statement.")
		}

	case lexer.TT_LT_Bool, lexer.TT_LT_Int, lexer.TT_LT_Float, lexer.TT_LT_String: // Literals
		startChar := sn.peek().Position.StartChar
		sn.analyzeExpression()
		sn.newErrorFromTo(sn.peek().Position.Line, startChar, sn.peek().Position.StartChar, "Expression can't be a statement.")

	case lexer.TT_KW_var, lexer.TT_KW_bool, lexer.TT_KW_int, lexer.TT_KW_flt, lexer.TT_KW_str: // Variable declarations
		sn.analyzeVariableDeclaration(false)

	case lexer.TT_KW_const: // Constant declarations
		sn.consume()
		sn.analyzeVariableDeclaration(true)

	case lexer.TT_DL_BraceClose: // Leave scope
		if isScope {
			return true
		}
		sn.newError(sn.peek(), fmt.Sprintf("Unexpected token \"%s\". Expected statement.", sn.consume()))

	case lexer.TT_DL_ParenthesisClose:
		return true

	case lexer.TT_DL_BraceOpen: // Enter scope
		sn.analyzeScope()

	case lexer.TT_KW_struct: // Struct
		sn.analyzeStructDefinition()

	case lexer.TT_KW_enum: // Enum
		sn.analyzeEnumDefinition()

	case lexer.TT_KW_if: // If
		sn.analyzeIfStatement(false)

	case lexer.TT_KW_else: // Else
		sn.newError(sn.peek(), "Else statement is missing an if statement.")
		sn.analyzeElseStatement()

	case lexer.TT_KW_loop: // Loop
		sn.analyzeLoop()

	case lexer.TT_KW_while: // While loop
		sn.analyzeWhileLoop()

	case lexer.TT_KW_for: // For loop
		sn.analyzeForLoop()

	case lexer.TT_KW_forEach: // ForEach loop
		sn.analyzeForEachLoop()

	case lexer.TT_KW_break: // Break
		sn.consume()

	case lexer.TT_KW_return: // Return
		sn.consume()

		if sn.peek().TokenType != lexer.TT_EndOfCommand {
			sn.analyzeExpression()
		}

	case lexer.TT_EndOfCommand: // Ignore EOCs

	default:
		// Collect line and print error
		startChar := sn.peek().Position.StartChar
		statement := sn.collectLine()
		sn.newErrorFromTo(sn.peek().Position.Line, startChar, sn.peek().Position.EndChar, fmt.Sprintf("Invalid statement \"%s\".", statement))
	}

	// Collect tokens after statement
	if sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_BraceClose {
		if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose {
			return true
		}

		startChar := sn.peek().Position.StartChar
		statement := sn.collectLine()
		sn.newErrorFromTo(sn.peek().Position.Line, startChar, sn.peek().Position.EndChar, fmt.Sprintf("Unexpected token/s \"%s\" after statement.", statement))
	}
	sn.consume()

	return false
}

func (sn *SyntaxAnalyzer) analyzeIdentifier() {
	// Enum variable declaration
	if sn.customTypes[sn.peek().Value] {
		sn.analyzeVariableDeclaration(false)
		return
	}

	sn.consume()

	// Function call
	if sn.peek().TokenType == lexer.TT_DL_ParenthesisOpen {
		sn.analyzeFunctionCall()
		// Assignment
	} else if sn.peek().TokenType.IsAssignKeyword() {
		sn.analyzeAssignment()
	}
}

func (sn *SyntaxAnalyzer) analyzeAssignment() {
	sn.consume()

	sn.analyzeExpression()
}
