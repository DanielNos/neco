package main

import (
	"flag"
	"fmt"

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
	// Define and collect flags
	tokens := flag.Bool("tokens", false, "Prints tokens when compiling.")
	tree := flag.Bool("tree", false, "Displays AST when compiling.")
	build := flag.String("build", "", "Builds specified source file.")

	flag.Parse()
	
	// Select action and target
	action := ""
	target := ""
	// Build
	if *build != "" {
		action = "build"
		target = *build
	// No action
	} else {
		Fatal(ERROR_INVALID_USE, "No action specified.")
	}
	return action, *tokens, *tree, target
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

	Info(fmt.Sprintf("Lexed %d tokens.\n", len(tokens)))

	// Analyze syntax
	syntaxAnalyzer := NewSyntaxAnalyzer(tokens)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.errorCount != 0 {
		Fatal(ERROR_SYNTAX, fmt.Sprintf("Syntax analysis failed with %d errors.", syntaxAnalyzer.errorCount + lexer.errorCount))
	}

	Success("Passed syntax analysis.")
}

func main() {
	action, showTokens, showTree, target := processArguments()

	if action == "build" {
		compile(target, showTokens, showTree)
	}
}
