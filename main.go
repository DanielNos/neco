package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/fatih/color"

	codeGen "neko/codeGenerator"
	"neko/errors"
	"neko/lexer"
	"neko/logger"
	"neko/parser"
	"neko/syntaxAnalyzer"
	VM "neko/virtualMachine"
)

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
			fmt.Printf("LINE_OFFSET              %d\n", instruction.InstructionType-128)
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

func processArguments() (string, bool, bool, bool, bool, string) {
	// Define and collect flags
	tokens := flag.Bool("tokens", false, "Prints tokens when compiling.")
	tree := flag.Bool("tree", false, "Displays AST when compiling.")
	instructions := flag.Bool("instructions", false, "Displays compiled instructions.")
	time := flag.Bool("time", false, "Shows execution time of program.")

	build := flag.String("build", "", "Builds specified source file.")
	run := flag.String("run", "", "Runs specified NeCo program.")

	flag.Parse()

	// Select action and target
	action := ""
	target := ""
	// Build
	if *build != "" {
		action = "build"
		target = *build
		// No action
	} else if *run != "" {
		action = "run"
		target = *run
	} else {
		logger.Fatal(errors.ERROR_INVALID_USE, "No action specified.")
	}
	return action, *tokens, *tree, *instructions, *time, target
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
	codeGenerator := codeGen.NewGenerator(tree, path[:len(path)-5])
	instructions := codeGenerator.Generate()

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
	action, showTokens, showTree, printInstruction, measureTime, target := processArguments()

	if action == "build" {
		compile(target, showTokens, showTree, printInstruction)
	} else if action == "run" {
		startTime := time.Now()

		virtualMachine := VM.NewVirutalMachine()
		virtualMachine.Execute(target)

		if measureTime {
			fmt.Printf("Execution time: %v\n", time.Since(startTime))
		}
	}
}
