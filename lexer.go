package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TokenType int8

const (
	TT_Identifier TokenType = iota
)

type Token struct {
	position *CodePos
	tokenType TokenType
	value string
}

type Lexer struct {
	filePath string
	file *os.File
	reader *bufio.Reader

	currRune rune
	nextRune rune

	tokens []Token
}

func NewLexer(filePath string) Lexer {
	return Lexer{filePath, nil, nil, ' ', ' ', make([]Token, 0, 100)}
}

func (l *Lexer) advance() {
	l.currRune = l.nextRune
	
	r, _, err := l.reader.ReadRune()
	if err != nil {
		l.nextRune = EOF
		l.file.Close()
	} else {
		l.nextRune = r
	}
}

func (l *Lexer) Lex() []Token {
	// Create reader
	file, err := os.Open(l.filePath)
	l.file = file

	if err != nil {
		fatal(2, fmt.Sprintf("Failed to open file %s: %s.", l.filePath, strings.Split(err.Error(), ": ")[1]))
	}

	// Read first 2 chars
	l.reader = bufio.NewReader(file)
	l.advance()
	l.advance()

	for {
		if l.currRune == EOF {
			return l.tokens
		}
		fmt.Printf("%c",l.currRune)
		l.advance()
	}
}
