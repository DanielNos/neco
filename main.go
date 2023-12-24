package main

import (
	"flag"
	"fmt"

	"github.com/fatih/color"

	"neko/errors"
	"neko/lexer"
	"neko/logger"
	"neko/parser"
	"neko/syntaxAnalyzer"
)

func printTokens(tokens []*lexer.Token) {
	for _, token := range tokens {
		if token.TokenType > lexer.TT_KW_const {
			color.Set(color.FgHiCyan)
		} else if token.TokenType == lexer.TT_EndOfCommand {
			color.Set(color.FgHiYellow)
		} else if token.TokenType == lexer.TT_Identifier {
			color.Set(color.FgHiGreen)
		} else if token.TokenType == lexer.TT_StartOfFile || token.TokenType == lexer.TT_EndOfFile {
			color.Set(color.FgHiRed)
		} else if token.TokenType >= lexer.TT_OP_Add && token.TokenType <= lexer.TT_OP_GreaterEqual {
			color.Set(color.FgHiMagenta)
		} else if token.TokenType >= lexer.TT_DL_ParenthesisOpen && token.TokenType <= lexer.TT_DL_Comma {
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
		logger.Fatal(errors.ERROR_INVALID_USE, "No action specified.")
	}
	return action, *tokens, *tree, target
}

func compile(path string, showTokens, showTree bool) {
	logger.Info(fmt.Sprintf("Compiling %s.\n", path))

	// Tokenize
	lexer := lexer.NewLexer(path)
	tokens := lexer.Lex()

	logger.Info(fmt.Sprintf("Lexed %d tokens.\n", len(tokens)))

	// Print tokens
	if showTokens {
		printTokens(tokens)
		println()
	}

	// Analyze syntax
	syntaxAnalyzer := syntaxAnalyzer.NewSyntaxAnalyzer(tokens, lexer.ErrorCount)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.ErrorCount != 0 {
		logger.Fatal(errors.ERROR_SYNTAX, fmt.Sprintf("Syntax analysis failed with %d errors.", syntaxAnalyzer.ErrorCount))
	}

	logger.Success("Passed syntax analysis.")

	// Construct AST
	p := parser.NewParser(tokens, syntaxAnalyzer.ErrorCount)
	tree := p.Parse()

	if p.ErrorCount != 0 {
		logger.Fatal(errors.ERROR_SEMANTIC, fmt.Sprintf("Semantic analysis failed with %d errors.", p.ErrorCount))
	}

	// Visualize tree
	if showTree {
		println()
		parser.Visualize(tree)
		println()
	}

	logger.Success("Passed semantic analysis.")
}

func main() {
	action, showTokens, showTree, target := processArguments()

	if action == "build" {
		compile(target, showTokens, showTree)
	}
}
