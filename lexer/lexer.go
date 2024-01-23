package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"

	"neco/dataStructures"
	"neco/errors"
	"neco/logger"
)

const EOF rune = 0x04

var TOKEN_BREAKERS = map[rune]struct{}{
	'+': {},
	'-': {},
	'*': {},
	'/': {},
	'^': {},
	'%': {},

	'=': {},
	'<': {},
	'>': {},

	'!': {},
	'&': {},
	'|': {},

	';': {},
}

func isTokenBreaker(char rune) bool {
	_, isBreaker := TOKEN_BREAKERS[char]
	_, isDelimiter := DELIMITERS[char]

	return unicode.IsSpace(char) || isBreaker || isDelimiter
}

type Lexer struct {
	filePath string
	file     *os.File
	reader   *bufio.Reader
	fileOpen bool

	currRune rune
	nextRune rune

	lineIndex uint
	charIndex uint

	token  bytes.Buffer
	tokens []*Token

	ErrorCount uint
}

func NewLexer(filePath string) Lexer {
	return Lexer{filePath, nil, nil, false, ' ', ' ', 0, 0, bytes.Buffer{}, make([]*Token, 0, 100), 0}
}

func (l *Lexer) Lex() []*Token {
	// Create reader
	file, err := os.Open(l.filePath)
	l.file = file
	l.fileOpen = true

	if err != nil {
		reason := strings.Split(err.Error(), ": ")[1]
		logger.Fatal(errors.LEXICAL, fmt.Sprintf("Failed to open file %s. %c%s.", l.filePath, unicode.ToUpper(rune(reason[0])), reason[1:]))
	}

	l.newTokenFrom(0, 0, TT_StartOfFile, l.filePath)

	// Read first 2 chars
	l.reader = bufio.NewReader(file)
	l.advance()
	l.advance()

	l.charIndex = 1
	l.lineIndex = 1

	for {
		l.lexRune()

		if l.currRune == EOF {
			l.newToken(l.lineIndex, l.charIndex, TT_EndOfFile)
			return l.tokens
		}
	}
}

func (l *Lexer) newError(line, char uint, useTokenLength bool, message string) {
	if l.ErrorCount == 0 {
		fmt.Fprintf(os.Stderr, "\n")
	}

	l.ErrorCount++

	var tokenLength uint = 1
	if useTokenLength {
		tokenLength = uint(l.token.Len())
	}
	logger.ErrorPos(&l.filePath, line, char, char+tokenLength, message)

	// Too many errors
	if l.ErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.SYNTAX, fmt.Sprintf("Lexical analysis has aborted due to too many errors. It has failed with %d errors.", l.ErrorCount))
	}
}

func (l *Lexer) advance() {
	// Don't advance when file is closed
	if !l.fileOpen {
		// Move EOF to current rune if it's not there yet
		if l.currRune != EOF {
			l.currRune = EOF
			l.charIndex++
		}
		return
	}

	// Move next to current
	l.currRune = l.nextRune

	r, _, err := l.reader.ReadRune()
	// Failed to read rune
	if err != nil {
		l.nextRune = EOF
		l.file.Close()
		l.fileOpen = false
		// Read rune
	} else {
		l.nextRune = r
	}
	l.charIndex++
}

func (l *Lexer) newToken(startLine, startChar uint, tokenType TokenType) {
	l.newTokenFrom(startLine, startChar, tokenType, l.token.String())
	l.token.Reset()
}

func (l *Lexer) newTokenFrom(startLine, startChar uint, tokenType TokenType, value string) {
	l.tokens = append(l.tokens, &Token{&dataStructures.CodePos{File: &l.filePath, Line: startLine, StartChar: startChar, EndChar: l.charIndex}, tokenType, value})
}

func (l *Lexer) collectRestOfToken() {
	for !unicode.IsSpace(l.currRune) && !isTokenBreaker(l.currRune) {
		l.token.WriteRune(l.currRune)
		l.advance()
	}
}

func (l *Lexer) lexRune() {
	if unicode.IsLetter(l.currRune) { // Identifier/Keyword
		l.lexLetter()
	} else if unicode.IsDigit(l.currRune) { // Int/Float
		l.lexNumber()
	} else {
		switch l.currRune {

		case '_': // Identifier
			l.lexLetter()

		case '"': // String
			l.lexString()

		case '\n': // New Line
			l.newTokenFrom(l.lineIndex, l.charIndex, TT_EndOfCommand, "")
			l.advance()

			l.charIndex = 1
			l.lineIndex++

		case '\r': // Windows New Line
			l.newTokenFrom(l.lineIndex, l.charIndex, TT_EndOfCommand, "")
			l.advance()

			// Invalid Windows line ending
			if l.currRune != '\n' {
				l.newError(l.lineIndex, l.charIndex-1, true, "Invalid Windows line ending.")
			} else {
				l.advance()
			}

		// Boolean operators
		case '=':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Equal, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_Assign, "")
			}
		case '<':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_LowerEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Lower, "")
			}
		case '>':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_GreaterEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Greater, "")
			}
		// Operators
		case '+':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_AddAssign, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Add, "")
			}
		case '-':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_SubtractAssign, "")
				l.advance()
			} else if l.currRune == '>' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_returns, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Subtract, "")
			}
		case '*':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_MultiplyAssign, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Multiply, "")
			}
		case '/':
			l.advance()
			if l.currRune == '=' { // /=
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_DivideAssign, "")
				l.advance()
			} else if l.currRune == '/' { // //
				l.advance()
				l.skipComment()
			} else if l.currRune == '*' { // /*
				l.advance()
				l.skipMultiLineComment()
			} else { // /
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Divide, "")
			}
		case '^':
			l.advance()
			if l.currRune == '=' { // ^=
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_PowerAssign, "")
				l.advance()
			} else { // ^
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Power, "")
			}
		case '%':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_KW_ModuloAssign, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Modulo, "")
			}
		case '!':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_NotEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Not, "")
			}
		case '&':
			l.advance()
			l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_And, "")
		case '|':
			l.advance()
			l.newTokenFrom(l.lineIndex, l.charIndex-1, TT_OP_Or, "")

		// EOC
		case ';':
			l.newTokenFrom(l.lineIndex, l.charIndex, TT_EndOfCommand, ";")
			l.advance()

		default:
			// Delimiters
			delimiter, isDelimiter := DELIMITERS[l.currRune]

			if isDelimiter {
				l.newTokenFrom(l.lineIndex, l.charIndex, delimiter, "")
			} else {
				// Invalid character
				if !unicode.IsSpace(l.currRune) && l.currRune != EOF {
					l.token.WriteRune(l.currRune)
					l.newError(l.lineIndex, l.charIndex, true, fmt.Sprintf("Invalid character \"%c\".", l.currRune))
				}
			}
			l.advance()
		}
	}
}
