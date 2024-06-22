package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"

	codeGen "neco/codeGenerator"
	"neco/errors"
	"neco/lexer"
	"neco/logger"
	"neco/parser"
	"neco/syntaxAnalyzer"
	VM "neco/virtualMachine"
)

type Action byte

const (
	A_Build Action = iota
	A_Run
	A_Analyze
)

type Configuration struct {
	PrintTokens       bool
	DrawTree          bool
	PrintInstructions bool
	Optimize          bool
	Silent            bool
	PrintConstants    bool

	Action     Action
	TargetPath string
	OutputPath string
}

func printHelp() {
	color.Set(color.Bold)
	color.Set(color.FgHiYellow)
	fmt.Println("Action           Flags")

	color.Set(color.Reset)
	fmt.Println("build [target]")
	fmt.Println("                 -to --tokens            Prints lexed tokens.")
	fmt.Println("                 -tr --tree              Draws abstract syntax tree.")
	fmt.Println("                 -i  --instructions      Prints generated instructions.")
	fmt.Println("                 -d  --dont-optimize      Compiler won't optimize byte code.")
	fmt.Println("                 -s  --silent            Doesn't produce info messages when possible.")
	fmt.Println("                 -n  --no-log             Doesn't produce any log messages, even if there are errors.")
	fmt.Println("                 -l  --log-level [LEVEL]  Sets logging level. Possible values are 0 to 5.")
	fmt.Println("                 -o  --out               Sets output file path.")
	fmt.Println("                 -c  --constants         Prints constants stored in binary.")
	fmt.Println("\nrun [target]")
	fmt.Println("\nanalyze [target]")
	fmt.Println("                 -to --tokens        Prints lexed tokens.")
	fmt.Println("                 -tr --tree          Draws abstract syntax tree.")
	fmt.Println("                 -d  --dontOptimize  Compiler won't optimize byte code.")
}

func processArguments() *Configuration {
	args := os.Args[1:]

	// No action
	if len(args) == 0 {
		logger.Fatal(errors.INVALID_FLAGS, "No action specified. Use neco help for more info.")
	}

	argumentsStart := 2

	// Collect target
	configuration := &Configuration{Optimize: true}

	switch args[0] {
	case "build", "run", "analyze":
		if len(args) == 1 {
			logger.Fatal(errors.INVALID_FLAGS, "No target specified.")
		}
		configuration.TargetPath = args[1]

		switch args[0] {
		case "build":
			configuration.Action = A_Build

		case "run":
			configuration.Action = A_Run

		case "analyze":
			configuration.Action = A_Analyze

		}

	case "help", "--help", "-h":
		printHelp()
		os.Exit(0)

	default:
		if strings.HasSuffix(args[0], ".neco") {
			configuration.Action = A_Build
		} else {
			configuration.Action = A_Run
		}

		configuration.TargetPath = args[0]
		argumentsStart = 1
	}

	// Collect flags
	switch configuration.Action {
	// Build flags
	case A_Build:
		for i := argumentsStart; i < len(args); i++ {
			switch args[i] {
			case "--tokens", "-to":
				configuration.PrintTokens = true

			case "--tree", "-tr":
				configuration.DrawTree = true

			case "--instructions", "-i":
				configuration.PrintInstructions = true

			case "--dont-optimize", "-d":
				configuration.Optimize = false

			case "--silent", "-s":
				logger.LoggingLevel = logger.LL_Error

			case "--no-log", "-n":
				logger.LoggingLevel = logger.LL_NoLog

			case "--log-level", "-l":
				if i+1 == len(args) {
					logger.Fatal(errors.INVALID_FLAGS, "No logging level provided after "+args[i]+" flag.")
				}

				i++
				loggingLevel, err := strconv.Atoi(args[i])

				if err != nil {
					logger.Fatal(errors.INVALID_FLAGS, "Logging level has to be a number.")
				}

				if loggingLevel < 0 || loggingLevel > 5 {
					logger.Fatal(errors.INVALID_FLAGS, "Invalid logging level "+fmt.Sprintf("%d.", loggingLevel))
				}

				logger.LoggingLevel = byte(loggingLevel)

			case "--out", "-o":
				if i+1 == len(args) {
					logger.Fatal(errors.INVALID_FLAGS, "No output path provided after "+args[i]+" flag.")
				}
				i++

				configuration.OutputPath = args[i]

			case "--constants", "-c":
				configuration.PrintConstants = true

			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+args[i]+"\" for action build.")
			}
		}
	// Run flags
	case A_Run:
		for _, flag := range args[argumentsStart:] {
			switch flag {
			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action run.")
			}
		}
	// Analyze flags
	case A_Analyze:
		for _, flag := range args[2:] {
			switch flag {
			case "--tokens", "-to":
				configuration.PrintTokens = true

			case "--tree", "-tr":
				configuration.DrawTree = true

			case "--dont-optimize", "-d":
				configuration.Optimize = false

			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action analyze.")
			}
		}
	}

	// Set output binary path
	if configuration.OutputPath == "" {
		if strings.HasSuffix(configuration.TargetPath, ".neco") && len(configuration.TargetPath) > 5 {
			configuration.OutputPath = configuration.TargetPath[:len(configuration.TargetPath)-5]

			if runtime.GOOS == "windows" {
				configuration.OutputPath += ".nc"
			}
		} else if runtime.GOOS == "windows" {
			configuration.OutputPath += ".nc"
		} else {
			configuration.OutputPath += "_bin"
		}
	}

	return configuration
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

	logger.Info(fmt.Sprintf("Lexed %d tokens.", len(tokens)))

	// Analyze syntax
	syntaxAnalyzer := syntaxAnalyzer.NewSyntaxAnalyzer(tokens, lexer.ErrorCount)
	syntaxAnalyzer.Analyze()

	if syntaxAnalyzer.ErrorCount != 0 {
		logger.Error(fmt.Sprintf("Syntax analysis failed with %d error/s.", syntaxAnalyzer.ErrorCount))

		// Print tokens
		if configuration.PrintTokens {
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
		printInstructions(&codeGenerator.GlobalsInstructions, codeGenerator.Constants, int(codeGenerator.FirstLine))
		printInstructions(&codeGenerator.FunctionsInstructions, codeGenerator.Constants, int(codeGenerator.FirstLine))

		fmt.Println()
	}
}

func main() {
	configuration := processArguments()

	// Build target
	if configuration.Action == A_Build {
		logger.Info("🐱 Compiling " + configuration.TargetPath)
		compile(configuration)
		// Run target
	} else if configuration.Action == A_Run {
		virtualMachine := VM.NewVirutalMachine(configuration.TargetPath)
		virtualMachine.Execute()
		// Analyze target
	} else if configuration.Action == A_Analyze {
		logger.Info("🐱 Analyzing " + configuration.TargetPath)
		startTime := time.Now()

		analyze(configuration)

		logger.Success(fmt.Sprintf("😺 Analyze completed in %s.", time.Since(startTime)))
	}
}
