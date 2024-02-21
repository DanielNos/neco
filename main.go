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
	println("Action           Flags")

	color.Set(color.Reset)
	println("build [target]")
	println("                 -to --tokens       Prints lexed tokens.")
	println("                 -tr --tree         Draws abstract syntax tree.")
	println("                 -i --instructions  Prints generated instructions.")
	println("                 -d --dontOptimize  Compiler won't optimize byte code.")
	println("\nrun [target]")
	println("\nanalyze [target] -to --tokens   Prints lexed tokens.")
	println("                 -tr --tree     Draws abstract syntax tree.")
}

func printTokens(tokens []*lexer.Token) {
	println()
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
	println()
}

func printInstructions(instructions *[]VM.Instruction) {
	line := int((*instructions)[0].InstructionValue[0]) - 1
	linePadder := "  "
	justChanged := true

	for i, instruction := range *instructions {
		// Skip removed instruction
		if instruction.InstructionType == 255 {
			continue
		}

		// Display empty line instead of offset and record new line number
		if instruction.InstructionType == VM.IT_LineOffset {
			if line < 10 && line+int(instruction.InstructionValue[0]) >= 10 {
				linePadder = linePadder[1:]
			}

			line += int(instruction.InstructionValue[0])
			justChanged = true

			println()
			continue
		}

		// Print line number
		if justChanged {
			fmt.Printf("%s%d  ", linePadder, line)
			justChanged = false
		} else {
			print("     ")
		}

		// Print instruction name
		if i < 10 {
			print(" ")
		}
		fmt.Printf("%d   %s", i, VM.InstructionTypeToString[instruction.InstructionType])

		i := len(VM.InstructionTypeToString[instruction.InstructionType])
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
		logger.Fatal(errors.INVALID_FLAGS, "No action specified. Use neco help for more info.")
	}

	action := args[0]

	// Collect target
	target := ""

	switch action {
	case "build", "run", "analyze":
		if len(args) == 1 {
			logger.Fatal(errors.INVALID_FLAGS, "No target specified.")
		}
		target = args[1]
	case "help":
		printHelp()
		os.Exit(0)
	default:
		logger.Fatal(errors.INVALID_FLAGS, fmt.Sprintf("Invalid action %s. Use neco help for more info.", args[1]))
	}

	// Collect flags
	var flags []bool

	switch action {
	// Build flags
	case "build":
		flags = []bool{false, false, false, false}
		for _, flag := range args[2:] {
			switch flag {
			case "--tokens", "-to":
				flags[0] = true
			case "--tree", "-tr":
				flags[1] = true
			case "--instructions", "-i":
				flags[2] = true
			case "--dontOptimize", "-d":
				flags[3] = true
			default:
				logger.Fatal(errors.INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action build.", flag))
			}
		}
	// Run flags
	case "run":
		for _, flag := range args[2:] {
			switch flag {
			default:
				logger.Fatal(errors.INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action run.", flag))
			}
		}
	// Analyze flags
	case "analyze":
		flags = []bool{false, false}
		for _, flag := range args[2:] {
			switch flag {
			case "--tokens", "-to":
				flags[0] = true
			case "--tree", "-tr":
				flags[1] = true
			default:
				logger.Fatal(errors.INVALID_FLAGS, fmt.Sprintf("Invalid flag \"%s\" for action analyze.", flag))
			}
		}
	}

	return action, target, flags
}

func analyze(path string, showTokens, showTree, isCompiling bool) (*parser.Node, *parser.Parser) {
	action := "Analysis"
	if isCompiling {
		action = "Compilation"
	}

	// Tokenize
	lexer := lexer.NewLexer(path)
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
		if showTokens {
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
	p := parser.NewParser(tokens, syntaxAnalyzer.ErrorCount)
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
	if showTokens {
		printTokens(tokens)
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
		logger.Fatal(exitCode, fmt.Sprintf("😿 %s failed with %d error/s.", action, lexer.ErrorCount+syntaxAnalyzer.ErrorCount+p.ErrorCount))
	}

	return tree, &p
}

func compile(path string, showTokens, showTree, printInstruction, dontOptimize bool) {
	startTime := time.Now()

	tree, p := analyze(path, showTokens, showTree, true)

	// Generate code
	codeGenerator := codeGen.NewGenerator(tree, path[:len(path)-5], p.IntConstants, p.FloatConstants, p.StringConstants, !dontOptimize)
	instructions := codeGenerator.Generate()

	// Generation failed
	if codeGenerator.ErrorCount != 0 {
		logger.Fatal(errors.CODE_GENERATION, fmt.Sprintf("Failed code generation with %d error/s.", codeGenerator.ErrorCount))
	}

	logger.Info(fmt.Sprintf("Generated %d instructions.", len(*instructions)))
	logger.Success(fmt.Sprintf("😺 Compilation completed in %s.", time.Since(startTime)))

	codeWriter := codeGen.NewCodeWriter(codeGenerator)
	codeWriter.Write()

	// Print generated instructions
	if printInstruction {
		printInstructions(instructions)
		println()
	}
}

func main() {
	action, target, flags := processArguments()

	// Build target
	if action == "build" {
		logger.Info(fmt.Sprintf("🐱 Compiling %s", target))
		compile(target, flags[0], flags[1], flags[2], flags[3])
		// Run target
	} else if action == "run" {
		virtualMachine := VM.NewVirutalMachine()
		virtualMachine.Execute(target)
		// Analyze target
	} else if action == "analyze" {
		logger.Info(fmt.Sprintf("🐱 Analyzing %s", target))
		startTime := time.Now()

		analyze(target, flags[0], flags[1], false)

		logger.Success(fmt.Sprintf("😺 Analyze completed in %s.", time.Since(startTime)))
	}
}
