package main

import (
	"flag"
	"fmt"
	"time"

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
	color.Set(color.FgHiWhite)
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
	logger.Info(fmt.Sprintf("Compiling %s.", path))
	startTime := time.Now()

	// Tokenize
	lexer := lexer.NewLexer(path)
	tokens := lexer.Lex()

	exitCode := 0
	if lexer.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Lexical analysis failed with %d error/s.", lexer.ErrorCount))
		exitCode = errors.ERROR_LEXICAL
	} else {
		logger.Success("Passed lexical analysis.")
	}

	logger.Info(fmt.Sprintf("Lexed %d tokens.", len(tokens)))

	// Analyze syntax
	syntaxAnalyzer := syntaxAnalyzer.NewSyntaxAnalyzer(tokens, lexer.ErrorCount)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Syntax analysis failed with %d error/s.", syntaxAnalyzer.ErrorCount))
		
		// Print tokens
		if showTokens {
			println()
			printTokens(tokens)
			println()
		}

		// Exit with correct return code
		if exitCode == 0 {
			logger.Fatal(errors.ERROR_SYNTAX, fmt.Sprintf("Compilation failed with %d error/s.", lexer.ErrorCount + syntaxAnalyzer.ErrorCount))
		} else {
			logger.Fatal(exitCode, fmt.Sprintf("Compilation failed with %d error/s.", lexer.ErrorCount + syntaxAnalyzer.ErrorCount))
		}
	} else {
		logger.Success("Passed syntax analysis.")
	}

	// Construct AST
	p := parser.NewParser(tokens, syntaxAnalyzer.ErrorCount)
	tree := p.Parse()

	// Print info
	if p.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Semantic analysis failed with %d error/s.", p.ErrorCount))
		if exitCode == 0 {
			exitCode = errors.ERROR_SEMANTIC
		}
	} else {
		logger.Success("Passed semantic analysis.")
	}

	// Print tokens
	if showTokens {
		println()
		printTokens(tokens)
		println()
	}

	// Visualize tree
	if showTree {
		if !showTokens {
			println()
		}
		parser.Visualize(tree)
		println()
	}

	if exitCode != 0 {
		logger.Fatal(exitCode, fmt.Sprintf("Compilation failed with %d error/s.", lexer.ErrorCount + syntaxAnalyzer.ErrorCount + p.ErrorCount))
	}

	logger.Success(fmt.Sprintf("Compilation completed in %s.", time.Since(startTime)))
}

func main() {
	action, showTokens, showTree, target := processArguments()

	if action == "build" {
		compile(target, showTokens, showTree)
	}
}
