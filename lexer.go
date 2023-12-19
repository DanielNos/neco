package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

const EOF rune = 0x04

var TOKEN_BREAKERS = map[rune]bool {
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'^': true,
	'%': true,
	
	'=': true,
	'<': true,
	'>': true,
	
	'!': true,
	'&': true,
	'|': true,
	
	';': true,
}

var DIGIT_VALUE = map[rune]int {
	'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
	'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	'a': 10, 'A': 10, 'b': 11, 'B': 11,
	'c': 12, 'C': 12, 'd': 13, 'D': 13,
	'e': 14, 'E': 14, 'f': 15, 'F': 15,
	'g': 16, 'G': 16, 'h': 17, 'H': 17,
	'i': 18, 'I': 18, 'j': 19, 'J': 19,
	'k': 20, 'K': 20, 'l': 21, 'L': 21,
	'm': 22, 'M': 22, 'n': 23, 'N': 23,
	'o': 24, 'O': 24, 'p': 25, 'P': 25,
	'q': 26, 'Q': 26, 'r': 27, 'R': 27,
	's': 28, 'S': 28, 't': 29, 'T': 29,
	'u': 30, 'U': 30, 'v': 31, 'V': 31,
	'w': 32, 'W': 32, 'x': 33, 'X': 33,
	'y': 34, 'Y': 34, 'z': 35, 'Z': 35,
}

func isTokenBreaker(char rune) bool {
	_, breaker := TOKEN_BREAKERS[char]
	_, delimiter := DELIMITERS[char]

	return unicode.IsSpace(char) || breaker || delimiter
}

type Lexer struct {
	filePath string
	file *os.File
	reader *bufio.Reader
	fileOpen bool

	currRune rune
	nextRune rune

	lineIndex uint
	charIndex uint

	token bytes.Buffer
	tokens []*Token

	errorCount uint
}

func NewLexer(filePath string) Lexer {
	return Lexer{filePath, nil, nil, false, ' ', ' ', 1, 1, bytes.Buffer{}, make([]*Token, 0, 100), 0}
}

func (l *Lexer) Lex() []*Token {
	// Create reader
	file, err := os.Open(l.filePath)
	l.file = file
	l.fileOpen = true
	
	if err != nil {
		reason := strings.Split(err.Error(), ": ")[1]
		Fatal(ERROR_LEXICAL, fmt.Sprintf("Failed to open file %s. %c%s.", l.filePath, unicode.ToUpper(rune(reason[0])), reason[1:]))
	}

	l.newTokenFrom(0, 0, TT_StartOfFile, l.filePath)
	
	// Read first 2 chars
	l.reader = bufio.NewReader(file)
	l.advance()
	l.advance()
	
	for {
		l.lexRune()
		
		if l.currRune == EOF {
			l.newToken(l.lineIndex, l.charIndex, TT_EndOfFile)
			return l.tokens
		}
	}
}

func (l *Lexer) newError(line, char uint, message string) {
	l.errorCount++
	ErrorPos(&l.filePath, line, char, char + uint(l.token.Len()), message)

	// Too many errors
	if l.errorCount > MAX_ERROR_COUNT {
		Fatal(ERROR_SYNTAX, fmt.Sprintf("Lexical analysis has aborted due to too many errors. It has failed with %d errors.", l.errorCount))
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
	l.tokens = append(l.tokens, &Token{&CodePos{&l.filePath, startLine, startChar, l.charIndex}, tokenType, value})
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
				l.newError(l.lineIndex, l.charIndex - 1, "Invalid Windows line ending.")
			} else {
				l.advance()
			}
		
		// Boolean operators
		case '=':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Equal, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_Assign, "")
			}
		case '<':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_LowerEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Lower, "")
			}
		case '>':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_GreaterEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Greater, "")
			}
		// Operators
		case '+':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_AddAssign, "")
				l.advance()
			} else {
					l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Add, "")
			}
		case '-':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_SubtractAssign, "")
				l.advance()
			} else if l.currRune == '>' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_returns, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Subtract, "")
			}
		case '*':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_MultiplyAssign, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Multiply, "")
			}
		case '/':
			l.advance()
			if l.currRune == '=' { // /=
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_DivideAssign, "")
				l.advance()
			} else if l.currRune == '/' { // //
				l.advance()
				l.skipComment()
			} else if l.currRune == '*' { // /*
				l.advance()
				l.skipMultiLineComment()
			} else { // /
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Divide, "")
			}
		case '^':
			l.advance()
			if l.currRune == '=' { // ^=
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_PowerAssign, "")
				l.advance()
			} else { // ^
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Power, "")
			}
		case '%':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_KW_ModuloAssign, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Modulo, "")
			}
		case '!':
			l.advance()
			if l.currRune == '=' {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_NotEqual, "")
				l.advance()
			} else {
				l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Not, "")
			}
		case '&':
			l.advance()
			l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_And, "")
		case '|':
			l.advance()
			l.newTokenFrom(l.lineIndex, l.charIndex - 1, TT_OP_Or, "")
									
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
				if !unicode.IsSpace(l.currRune) {
					l.token.WriteRune(l.currRune)
					l.newError(l.lineIndex, l.charIndex, fmt.Sprintf("Invalid character \"%c\".", l.currRune))
				}
			}
			l.advance()
		}					
	}
}
							
func (l *Lexer) lexLetter() {
	startLine := l.lineIndex
	startChar := l.charIndex
	
	// Collect identifier/keyword
	l.token.WriteRune(l.currRune)
	l.advance()

	for unicode.IsLetter(l.currRune) || unicode.IsDigit(l.currRune) || l.currRune == '_' || l.currRune == '.' {
		l.token.WriteRune(l.currRune)
		l.advance()
	}

	// Check if token is a keyword
	value := l.token.String()
	keyword, isKeyword := KEYWORDS[value]

	// Identifier/Literal
	if !isKeyword {
		if value == "true" { // Literal true
			l.newTokenFrom(startLine, startChar, TT_LT_Bool, "1")
			l.token.Reset()
		} else if value == "false" { // Literal false
			l.newTokenFrom(startLine, startChar, TT_LT_Bool, "0")
			l.token.Reset()
		} else if value == "none" { // Literal none
			l.newTokenFrom(startLine, startChar, TT_LT_None, "")
			l.token.Reset()
		} else { // Identifier
			l.newToken(startLine, startChar, TT_Identifier)
		}
	// Keyword
	} else {
		l.newTokenFrom(startLine, startChar, keyword, "")
		l.token.Reset()
	}
}

func (l *Lexer) lexString() {
	startLine := l.lineIndex
	startChar := l.charIndex
	l.advance()

	// Collect string
	for l.currRune != '"' {
		l.token.WriteRune(l.currRune)
		l.advance()
	}
	l.advance()

	l.newToken(startLine, startChar, TT_LT_String)
}

func (l *Lexer) lexNumber() {
	startLine := l.lineIndex
	startChar := l.charIndex
	
	// Collect number/base
	var base string

	for i := 0; i < 2; i++{
		if !unicode.IsDigit(l.currRune) && l.currRune != '_' && l.currRune != 'x' {
			l.collectRestOfToken()
			l.newError(startLine, startChar, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
			l.newToken(startLine, startChar, TT_LT_Int)
			return
		}

		l.token.WriteRune(l.currRune)
		l.advance()

		if l.currRune == 'x' {
			base = l.token.String()
			l.token.Reset()
			l.advance()
			break
		}

		if isTokenBreaker(l.currRune) {
			l.newToken(startLine, startChar, TT_LT_Int)
			return
		}
	}

	if base != "" {
		l.lexBaseInt(startLine, startChar, base)
		return
	}

	// Collect number
	for unicode.IsDigit(l.currRune) || l.currRune == '_' {
		if l.currRune != '_' {
			l.token.WriteRune(l.currRune)
		}
		l.advance()
	}
	
	// Create token
	if isTokenBreaker(l.currRune) {
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	// Float
	} else if l.currRune == '.' {
		l.lexFloat(startLine, startChar)
		return
	// Invalid characters in number
	} else {
		l.collectRestOfToken()
		l.newError(startLine, startChar, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
	}
}

func (l *Lexer) lexBaseInt(startLine, startChar uint, baseString string) {
	// Convert base to int
	base, _ := strconv.Atoi(baseString)

	// Invalid base
	if base < 2 || base > 36 {
		l.collectRestOfToken()
		l.newError(startLine, startChar, fmt.Sprintf("Invalid integer base %d. Only bases in range <2, 36> are supported.", base))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Collect number
	digitValue, valid := DIGIT_VALUE[l.currRune]
	invalidDigits := false
	for valid || l.currRune == '_' {
		if l.currRune != '_' {
			if digitValue >= base {
				invalidDigits = true
			}
			l.token.WriteRune(unicode.ToLower(l.currRune))
		}
		l.advance()
		digitValue, valid = DIGIT_VALUE[l.currRune]
	}

	// Digits exceed base
	if invalidDigits {
		l.newError(startLine, startChar + uint(len(baseString)) + 1, fmt.Sprintf("Digit/s of integer \"%s\" exceed its base.", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Invalid characters in number
	if !isTokenBreaker(l.currRune) {
		l.collectRestOfToken()
		l.newError(startLine, startChar, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	// Convert and create token
	value, _ := strconv.ParseInt(l.token.String(), base, 64)
	l.token.Reset()

	l.newTokenFrom(startLine, startChar, TT_LT_Int, fmt.Sprintf("%d", value))
}

func (l *Lexer) lexFloat(startLine, startChar uint) {
	l.token.WriteRune(l.currRune)
	l.advance()

	// Collect rest of float
	for unicode.IsDigit(l.currRune) || l.currRune == '_' {
		if l.currRune != '_' {
			l.token.WriteRune(l.currRune)
		}
		l.advance()
	}

	// Invalid characters
	if !isTokenBreaker(l.currRune) {
		l.collectRestOfToken()
		l.newError(startLine, startChar, fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	}

	l.newToken(startLine, startChar, TT_LT_Float)
}

func (l *Lexer) skipComment() {
	for l.currRune != '\n' && l.currRune != 'r' {
		l.advance()
	}
}

func (l *Lexer) skipMultiLineComment() {
	for l.currRune != EOF {
		switch l.currRune {

		case '*': // End of comment
			l.advance()
			if l.currRune == '/' {
				l.advance()
				return
			}

		case '/': // Start of new multiline comment
			l.advance()
			if l.currRune == '*' {
				l.advance()
				l.skipMultiLineComment()
			}

		case '\n': // New line
			l.lineIndex++;
			l.charIndex = 1;
			l.advance()

		case '\r': // Windows new line
			l.advance()
			if l.currRune != '\n' {
				l.newError(l.lineIndex, l.charIndex - 1, "Invalid Windows line ending.")
			} else {
				l.advance()
			}

			l.lineIndex++;
			l.charIndex = 1;

		default:
			l.advance()
		}
	}
}
