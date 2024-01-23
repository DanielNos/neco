package main

import (
	"fmt"
	"os"
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

func printHelp() {
	color.Set(color.Bold)
	color.Set(color.FgHiYellow)
	color.Set(color.ResetUnderline)
	println("Action           Flags")

	color.Set(color.Reset)
	println("build [target]")
	println("                 -tokens        Prints lexed tokens.")
	println("                 -tree          Draws abstract syntax tree.")
	println("                 -instructions  Prints generated instructions.")
	println("\nrun [target]     -time          Measures execution time.")
	println("\nanalyze [target]")
}

func printTokens(tokens []*lexer.Token) {
	for _, token := range tokens {
		if token.TokenType >= lexer.TT_KW_const {
			color.Set(color.FgHiCyan)
		} else if token.TokenType == lexer.TT_EndOfCommand {
			color.Set(color.FgHiYellow)
		} else if token.TokenType == lexer.TT_Identifier {
			color.Set(color.FgHiGreen)
		} else if token.TokenType == lexer.TT_StartOfFile || token.TokenType == lexer.TT_EndOfFile {
			color.Set(color.FgHiRed)
		} else if token.TokenType.IsOperator() {
			color.Set(color.FgHiMagenta)
		} else if token.TokenType.IsDelimiter() {
			color.Set(color.FgHiBlue)
		} else {
			color.Set(color.FgHiWhite)
		}
		fmt.Printf("%v\n", token.TableString())
	}
	color.Set(color.FgHiWhite)
}

func printInstructions(instructions *[]VM.Instruction) {
	for _, instruction := range *instructions {
		if instruction.InstructionType >= 128 {
			fmt.Printf("LINE_OFFSET              %d\n", instruction.InstructionType-128+1)
			continue
		}

		// Print instruction name
		fmt.Printf("%s", VM.InstructionTypeToString[instruction.InstructionType])

		i := len(fmt.Sprintf("%s", VM.InstructionTypeToString[instruction.InstructionType]))
		for i < 25 {
			print(" ")
			i++
		}

		// Print arguments
		for i := 0; i < len(instruction.InstructionValue); i++ {
			fmt.Printf("%d", instruction.InstructionValue[i])

			j := len(fmt.Sprintf("%d", instruction.InstructionValue[i]))
			for j < 5 {
				print(" ")
				j++
			}
		}

		println()
	}
}

func processArguments() (string, string, []bool) {
	args := os.Args[1:]

	// No action
	if len(args) == 0 {
		logger.Fatal(errors.ERROR_INVALID_FLAGS, "No action specified. Use neco help for more info.")
	}

	action := args[0]

	// Collect target
	target := ""

	switch action {
	case "build", "run", "analyze":
		if len(args) == 1 {
			logger.Fatal(errors.ERROR_INVALID_FLAGS, "No target specified.")
		}
		target = args[1]
	case "help":
		printHelp()
		os.Exit(0)
	default:
		logger.Fatal(errors.ERROR_INVALID_FLAGS, fmt.Sprintf("Invalid action %s. Use neco help for more info.", args[1]))
	}

	// Collect flags
	var flags []bool

	switch action {
	// Build flags
	case "build":
		flags = []bool{false, false, false}
		for _, flag := range args[2:] {
			switch flag {
			case "-tokens":
				flags[0] = true
			case "-tree":
				flags[1] = true
			case "-instructions":
				flags[2] = true
			default:
				logger.Fatal(errors.ERROR_INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action build.", flag))
			}
		}
	// Run flags
	case "run":
		flags = []bool{false}
		for _, flag := range args[2:] {
			switch flag {
			case "-time":
				flags[0] = true
			default:
				logger.Fatal(errors.ERROR_INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action run.", flag))
			}
		}
	// Analyze flags
	case "analyze":
		flags = []bool{}
		for _, flag := range args[2:] {
			switch flag {
			default:
				logger.Fatal(errors.ERROR_INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action analyze.", flag))
			}
		}
	}

	return action, target, flags
}

func compile(path string, showTokens, showTree, printInstruction bool) {
	logger.Info(fmt.Sprintf("🐱 Compiling %s", path))
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
			logger.Fatal(errors.ERROR_SYNTAX, fmt.Sprintf("😿 Compilation failed with %d error/s.", lexer.ErrorCount+syntaxAnalyzer.ErrorCount))
		} else {
			logger.Fatal(exitCode, fmt.Sprintf("😿 Compilation failed with %d error/s.", lexer.ErrorCount+syntaxAnalyzer.ErrorCount))
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
		logger.Fatal(exitCode, fmt.Sprintf("😿 Compilation failed with %d error/s.", lexer.ErrorCount+syntaxAnalyzer.ErrorCount+p.ErrorCount))
	}

	// Generate code
	codeGenerator := codeGen.NewGenerator(tree, path[:len(path)-5], p.IntConstants, p.FloatConstants, p.StringConstants)
	instructions := codeGenerator.Generate()

	// Generation failed
	if codeGenerator.ErrorCount != 0 {
		logger.Fatal(errors.ERROR_CODE_GENERATION, fmt.Sprintf("Failed code generation with %d error/s.", codeGenerator.ErrorCount))
	}

	logger.Info(fmt.Sprintf("Generated %d instructions.", len(*instructions)))
	logger.Success(fmt.Sprintf("😺 Compilation completed in %s.", time.Since(startTime)))

	codeWriter := codeGen.NewCodeWriter(codeGenerator)
	codeWriter.Write()

	// Print generated instructions
	if printInstruction {
		println()
		printInstructions(instructions)
		println()
	}
}

func main() {
	action, target, flags := processArguments()

	// Build target
	if action == "build" {
		compile(target, flags[0], flags[1], flags[2])
		// Run target
	} else if action == "run" {
		startTime := time.Now()

		virtualMachine := VM.NewVirutalMachine()
		virtualMachine.Execute(target)

		if flags[0] {
			fmt.Printf("Execution time: %v\n", time.Since(startTime))
		}
		// Analyze target
	} else if action == "analyze" {

	}
}
