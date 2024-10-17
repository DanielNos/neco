package main

import (
	"fmt"
	"time"

	"github.com/fatih/color"

	codeGen "github.com/DanielNos/neco/codeGenerator"
	"github.com/DanielNos/neco/errors"
	"github.com/DanielNos/neco/lexer"
	"github.com/DanielNos/neco/logger"
	"github.com/DanielNos/neco/parser"
	"github.com/DanielNos/neco/syntaxAnalyzer"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func printHelp() {
	color.Set(color.Bold)
	color.Set(color.FgHiYellow)
	fmt.Println("Action           Flags")

	color.Set(color.Reset)
	fmt.Println("build [target]")
	fmt.Println("                 -to --tokens            Prints lexed tokens.")
	fmt.Println("                 -tr --tree              Draws abstract syntax tree.")
	fmt.Println("                 -i  --instructions      Prints generated instructions.")
	fmt.Println("                 -d  --dont-optimize     Compiler won't optimize byte code.")
	fmt.Println("                 -s  --silent            Doesn't produce info messages when possible.")
	fmt.Println("                 -n  --no-log            Doesn't produce any log messages, even if there are errors.")
	fmt.Println("                 -l  --log-level [LEVEL] Sets logging level. Possible values are 0 to 5 or level names.")
	fmt.Println("                 -o  --out               Sets output file path.")
	fmt.Println("                 -c  --constants         Prints constants stored in binary.")
	fmt.Println("\nrun [target]")
	fmt.Println("\nanalyze [target]")
	fmt.Println("                 -to --tokens        Prints lexed tokens.")
	fmt.Println("                 -tr --tree          Draws abstract syntax tree.")
	fmt.Println("                 -d  --dontOptimize  Compiler won't optimize byte code.")
}

func analyze(configuration *Configuration) (*parser.Node, *parser.Parser) {
	action := "Analysis"
	if configuration.Action == A_Build {
		action = "Compilation"
	}

	// Tokenize
	lexer := lexer.NewLexer(configuration.TargetPath)
	tokens := lexer.Lex()

	exitCode := 0
	if lexer.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Lexical analysis failed with %d error/s.", lexer.ErrorCount))
		exitCode = errors.LEXICAL
	} else {
		logger.Success("Passed lexical analysis.")
	}

	// Analyze syntax
	syntaxAnalyzer := syntaxAnalyzer.NewSyntaxAnalyzer(tokens, lexer.ErrorCount)
	tokens = syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Syntax analysis failed with %d error/s.", syntaxAnalyzer.ErrorCount))

		// Print tokens
		if configuration.PrintTokens {
			logger.Info(fmt.Sprintf("Lexed %d tokens.", len(tokens)))
			printTokens(tokens)
		}

		// Exit with correct return code
		if exitCode == 0 {
			logger.Fatal(errors.SYNTAX, fmt.Sprintf("😿 %s failed with %d error/s.", action, lexer.ErrorCount+syntaxAnalyzer.ErrorCount))
		} else {
			logger.Fatal(exitCode, fmt.Sprintf("😿 %s failed with %d error/s.", action, lexer.ErrorCount+syntaxAnalyzer.ErrorCount))
		}
	} else {
		logger.Success("Passed syntax analysis.")
	}

	// Construct AST
	p := parser.NewParser(tokens, syntaxAnalyzer.ErrorCount, configuration.Optimize)
	tree := p.Parse()

	// Print info
	if p.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Semantic analysis failed with %d error/s.", p.ErrorCount))
		if exitCode == 0 {
			exitCode = errors.SEMANTIC
		}
	} else {
		logger.Success("Passed semantic analysis.")
	}

	// Print tokens
	if configuration.PrintTokens {
		printTokens(tokens)
	}

	// Visualize tree
	if configuration.DrawTree {
		if !configuration.PrintTokens {
			fmt.Println()
		}
		parser.Visualize(tree)
		fmt.Println()
	}

	if exitCode != 0 {
		logger.Fatal(exitCode, fmt.Sprintf("😿 %s failed with %d error/s.", action, lexer.ErrorCount+syntaxAnalyzer.ErrorCount+p.ErrorCount))
	}

	return tree, &p
}

func compile(configuration *Configuration) {
	if !configuration.Optimize {
		logger.Warning("Code optimization disabled.")
	}

	startTime := time.Now()

	tree, p := analyze(configuration)

	// Generate code
	codeGenerator := codeGen.NewGenerator(tree, p.IntConstants, p.FloatConstants, p.StringConstants, configuration.Optimize)
	codeGenerator.Generate()

	// Print constants
	if configuration.PrintConstants {
		printConstants(len(p.StringConstants), len(p.IntConstants), len(p.FloatConstants), codeGenerator.Constants)
	}

	// Generation failed
	if codeGenerator.ErrorCount != 0 {
		logger.Fatal(errors.CODE_GENERATION, fmt.Sprintf("Failed code generation with %d error/s.", codeGenerator.ErrorCount))
	}

	logger.Info(fmt.Sprintf("Generated %d instructions.", len(codeGenerator.GlobalsInstructions)+len(codeGenerator.FunctionsInstructions)))
	logger.Success(fmt.Sprintf("😺 Compilation completed in %s.", time.Since(startTime)))

	codeWriter := codeGen.NewCodeWriter(codeGenerator)
	codeWriter.Write(configuration.OutputPath)

	// Print generated instructions
	if configuration.PrintInstructions {
		printInstructions(&codeGenerator.GlobalsInstructions, codeGenerator.Constants)
		printInstructions(&codeGenerator.FunctionsInstructions, codeGenerator.Constants)

		fmt.Println()
	}
}

func buildAndRun(configuration *Configuration) {
	logger.Info("🐱 Compiling " + configuration.TargetPath)
	compile(configuration)

	virtualMachine := VM.NewVirtualMachine(configuration.OutputPath)
	virtualMachine.Execute()
}

func main() {
	configuration := processArguments()

	switch configuration.Action {
	case A_Build:
		logger.Info("🐱 Compiling " + configuration.TargetPath)
		compile(configuration)

	case A_Run:
		virtualMachine := VM.NewVirtualMachine(configuration.TargetPath)
		virtualMachine.Execute()

	case A_Analyze:
		logger.Info("🐱 Analyzing " + configuration.TargetPath)
		startTime := time.Now()

		analyze(configuration)

		logger.Success(fmt.Sprintf("😺 Analyze completed in %s.", time.Since(startTime)))

	case A_BuildAndRun:
		buildAndRun(configuration)
	}
}
