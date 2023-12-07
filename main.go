package main

import (
	"fmt"
	"os"
)

func compile(path string) {
	lexer := NewLexer(path)
	tokens := lexer.Lex()

	if lexer.errorCount != 0 {
		fatal(ERROR_LEXICAL, fmt.Sprintf("Lexical analysis failed with %d errors.", lexer.errorCount))
	}

	fmt.Printf("Lexed %d tokens.\n", len(tokens))
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fatal(ERROR_INVALID_USE, "No action or target specified.")
	}

	if len(args) == 1 {
		compile(args[0])
	} else {
		fatal(ERROR_INVALID_USE, "Invalid flags.")
	}	
}
