package main

import (
	"os"
)

func compile(path string) {
	lexer := NewLexer(path)
	lexer.Lex()
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fatal(1, "No action or target specified.")
	}

	if len(args) == 1 {
		compile(args[0])
	} else {
		fatal(1, "Invalid flags.")
	}	
}
