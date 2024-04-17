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
		fmt.Fprint(os.Stderr, "\n")
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
		fmt.Fprint(os.Stderr, "\n")
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

		expression += " " + sn.tokens[i].String()
		i++
	}

	// Return expression string
	if len(expression) > 0 {
		return expression[1:]
	} else {
		return ""
	}
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
				sn.newError(sn.peek(), "Too many EOCs (\\n or ;) after "+afterWhat+". Only 0 or 1 EOCs are allowed.")
			} else {
				sn.newError(sn.peek(), "Expected "+name+" after "+afterWhat+".")
				sn.rewind()
				return false
			}
		} else if sn.peek().TokenType != tokenType {
			if !optional {
				sn.newError(sn.peek(), "Expected "+name+" after "+afterWhat+".")
				sn.rewind()
				return false
			}
		}
	}

	return true
}

func (sn *SyntaxAnalyzer) consumeEOCs() {
	for sn.peek().TokenType == lexer.TT_EndOfCommand {
		sn.consume()
	}
}

func (sn *SyntaxAnalyzer) analyzeStatement(isScope bool) bool {
	if sn.peek().TokenType.IsVariableType() {
		sn.analyzeVariableDeclaration(false)
	}

	switch sn.peek().TokenType {

	case lexer.TT_KW_fun: // Function declaration
		sn.analyzeFunctionDeclaration()

	case lexer.TT_Identifier: // Identifiers
		sn.analyzeIdentifierStatement()

	case lexer.TT_LT_Bool, lexer.TT_LT_Int, lexer.TT_LT_Float, lexer.TT_LT_String: // Literals
		startChar := sn.peek().Position.StartChar
		sn.analyzeExpression()
		sn.newErrorFromTo(sn.peek().Position.StartLine, startChar, sn.peek().Position.StartChar, "Expression can't be a statement.")

	case lexer.TT_KW_const: // Constant declarations
		sn.consume()
		sn.analyzeVariableDeclaration(true)

	case lexer.TT_DL_BraceClose: // Leave scope
		if isScope {
			return true
		}
		sn.newError(sn.peek(), "Unexpected token \""+sn.consume().String()+"\". Expected statement.")

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
		return false

	case lexer.TT_KW_while: // While loop
		sn.analyzeWhileLoop()
		return false

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

	case lexer.TT_KW_delete: // Delete
		sn.consume()
		sn.analyzeExpression()

	case lexer.TT_KW_match: // Match
		sn.analyzeMatchStatement()

	case lexer.TT_KW_default:
		return false

	case lexer.TT_EndOfCommand: // Ignore EOCs
		sn.consume()
		return false

	default:
		// Collect rest of line and print error
		startChar := sn.peek().Position.StartChar

		for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_DL_BraceOpen && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.consume()
		}

		sn.newErrorFromTo(sn.peek().Position.StartLine, startChar, sn.peekPrevious().Position.EndChar, "Invalid statement.")
	}

	// Remaining tokens after statement
	if sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_EndOfFile && sn.peek().TokenType != lexer.TT_DL_BraceOpen {
		if sn.peek().TokenType == lexer.TT_DL_BraceClose && isScope {
			return true
		}

		if sn.peek().TokenType == lexer.TT_DL_ParenthesisClose || sn.peek().TokenType == lexer.TT_DL_BracketClose {
			sn.consume()
			return true
		}

		// Collect toklens after statement
		startPosition := sn.peek().Position

		for sn.peek().TokenType != lexer.TT_EndOfCommand && sn.peek().TokenType != lexer.TT_DL_BraceOpen && sn.peek().TokenType != lexer.TT_DL_ParenthesisClose {
			sn.consume()
		}

		sn.newErrorFromTo(startPosition.StartLine, startPosition.StartChar, sn.peekPrevious().Position.EndChar, "Unexpected token/s after statement.")
	}
	sn.consume()

	return false
}
