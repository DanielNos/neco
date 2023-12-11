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
		} else if token.tokenType == TT_StartOfFile || token.tokenType == TT_EndOfFile {
			color.Set(color.FgHiRed)
		} else if token.tokenType >= TT_OP_Add && token.tokenType <= TT_OP_GreaterEqual {
			color.Set(color.FgHiMagenta)
		} else if token.tokenType >= TT_DL_ParenthesisOpen && token.tokenType <= TT_DL_Comma {
			color.Set(color.FgHiBlue)
		} else {
			color.Set(color.FgHiWhite)
		}
		fmt.Printf("%v\n", token.TableString())
	}
}

func compile(path string) {
	lexer := NewLexer(path)
	tokens := lexer.Lex()

	printTokens(tokens)
	color.White("\n")

	info(fmt.Sprintf("Lexed %d tokens.", len(tokens)))

	syntaxAnalyzer := NewSyntaxAnalyzer(tokens)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.errorCount != 0 {
		fatal(ERROR_SYNTAX, fmt.Sprintf("Syntax analysis failed with %d errors.", syntaxAnalyzer.errorCount + lexer.errorCount))
	}

	info("Passed syntax analysis.")
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
