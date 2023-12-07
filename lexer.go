package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"unicode"
)

const EOF rune = 0x04

var TOKEN_BREAKERS = map[rune]bool {
	'+': true,
	'-': true,
	'*': true,
	'/': true,
	'%': true,

	'=': true,
}

func isTokenBreaker(char rune) bool {
	return TOKEN_BREAKERS[char] || unicode.IsSpace(char)
}

type Lexer struct {
	filePath string
	file *os.File
	reader *bufio.Reader

	currRune rune
	nextRune rune

	lineIndex uint
	charIndex uint

	token bytes.Buffer
	tokens []*Token

	errorCount uint
}

func NewLexer(filePath string) Lexer {
	return Lexer{filePath, nil, nil, ' ', ' ', 1, 1,  bytes.Buffer{}, make([]*Token, 0, 100), 0}
}

func (l *Lexer) Lex() []*Token {
	// Create reader
	file, err := os.Open(l.filePath)
	l.file = file
	
	if err != nil {
		fatal(ERROR_LEXICAL, fmt.Sprintf("Failed to open file %s: %s.", l.filePath, strings.Split(err.Error(), ": ")[1]))
	}
	
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

func (l *Lexer) newError(message string) {
	l.errorCount++
	error(message)
}

func (l *Lexer) advance() {
	l.currRune = l.nextRune
	
	r, _, err := l.reader.ReadRune()
	if err != nil {
		l.nextRune = EOF
		l.file.Close()
	} else {
		l.nextRune = r
		l.charIndex++
	}
}

func (l *Lexer) newToken(startLine, startChar uint, tokenType TokenType) {
	l.newTokenFrom(startLine, startChar, tokenType, l.token.String())
	l.token.Reset()
}

func (l *Lexer) newTokenFrom(startLine, startChar uint, tokenType TokenType, value string) {
	l.tokens = append(l.tokens, &Token{&CodePos{&l.filePath, startLine, l.lineIndex, startChar, l.charIndex}, tokenType, value})
	fmt.Printf("%v\n", l.tokens[len(l.tokens)-1])
}

func (l *Lexer) newEOCToken() {
	if l.tokens[len(l.tokens)-1].tokenType != TT_EndOfCommand {
		l.newToken(l.lineIndex, l.charIndex, TT_EndOfCommand)
	}
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
	} else if l.currRune == '"' { // String
		l.lexString()
	} else if unicode.IsDigit(l.currRune) { // Int/Float
		l.lexNumber()
	} else if l.currRune == '\n' { // New Line
		l.newEOCToken()
		l.advance()

		l.charIndex = 1
		l.lineIndex++
	} else if l.currRune == '\r' { // Windows New Line
		l.advance()

		// Invalid Windows line ending
		if l.currRune != '\n' {
			l.newError("Invalid Windows line ending.")
		} else {
			l.advance()
		}

	} else {
		l.advance()
	}
}

func (l *Lexer) lexLetter() {
	startLine := l.lineIndex
	startChar := l.charIndex

	l.token.WriteRune(l.currRune)
	l.advance()

	for unicode.IsLetter(l.currRune) || unicode.IsDigit(l.currRune) || l.currRune == '_' {
		l.token.WriteRune(l.currRune)
		l.advance()
	}

	value := l.token.String()
	keyword, exists := KEYWORDS[value]
	
	if !exists {
		l.newToken(startLine, startChar, TT_Identifier)
	} else {
		l.newTokenFrom(startLine, startChar, keyword, "")
		l.token.Reset()
	}
}

func (l *Lexer) lexString() {
	startLine := l.lineIndex
	startChar := l.charIndex
	l.advance()

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

	l.token.WriteRune(l.currRune)
	l.advance()

	for unicode.IsDigit(l.currRune) || l.currRune == '_' {
		l.token.WriteRune(l.currRune)
		l.advance()
	}
	
	if unicode.IsSpace(l.currRune) {
		l.newToken(startLine, startChar, TT_LT_Int)
		return
	} else {
		l.collectRestOfToken()
		l.newError(fmt.Sprintf("Invalid character/s in integer literal \"%s\".", l.token.String()))
	}

}
