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

func processArguments() (string, bool, bool, string) {
	args := os.Args[1:]

	// No args
	if len(args) == 0 {
		fatal(ERROR_INVALID_USE, "No action specified.")
	}

	// Collect action
	if args[0] != "build" {
		fatal(ERROR_INVALID_USE, fmt.Sprintf("Invalid action %s.", args[0]))
	}
	action := args[0]

	// No target
	if len(args) < 2 { 
		fatal(ERROR_INVALID_USE, "No target specified.")
	}

	// Collect flags
	tokens, tree := false, false
	
	for _, arg := range os.Args[2:len(args)] {
		if arg[0] == '-' && arg[1] == '-' {
			switch arg {
			case "--tokens":
				tokens = true
			case "--tree":
				tree = true
			default:
				fatal(ERROR_INVALID_USE, fmt.Sprintf("Invalid option %s.", arg))
			}
		}
	}
	
	return action, tokens, tree, args[len(args)-1]
}

func compile(path string, showTokens, showTree bool) {
	// Tokenize
	lexer := NewLexer(path)
	tokens := lexer.Lex()

	// Print tokens
	if showTokens {
		printTokens(tokens)
		println()
	}

	info(fmt.Sprintf("Lexed %d tokens.", len(tokens)))

	// Analyze syntax
	syntaxAnalyzer := NewSyntaxAnalyzer(tokens)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.errorCount != 0 {
		fatal(ERROR_SYNTAX, fmt.Sprintf("Syntax analysis failed with %d errors.", syntaxAnalyzer.errorCount + lexer.errorCount))
	}

	success("Passed syntax analysis.")
}

func main() {
	action, showTokens, showTree, target := processArguments()

	if action == "build" {
		compile(target, showTokens, showTree)
	}
}
