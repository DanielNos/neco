package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

func printTokens(tokens []*Token) {
	for _, token := range tokens {
		if token.tokenType > TT_KW_const {
			color.Set(color.FgHiCyan)
		} else if token.tokenType == TT_EndOfCommand {
			color.Set(color.FgHiYellow)
		} else if token.tokenType == TT_Identifier {
			color.Set(color.FgHiGreen)
		} else if token.tokenType == TT_EndOfFile {
			color.Set(color.FgHiRed)
		} else if token.tokenType > TT_OP_Add && token.tokenType < TT_OP_GreaterEqual {
			color.Set(color.FgHiMagenta)
		} else {
			color.Set(color.FgHiWhite)
		}
		fmt.Printf("%v\n", token)
	}
}

func compile(path string) {
	lexer := NewLexer(path)
	tokens := lexer.Lex()

	printTokens(tokens)
	color.White("\n")

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
