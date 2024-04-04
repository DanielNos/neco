package main

import (
	"fmt"
	"os"
	"reflect"
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

	Action     Action
	TargetPath string
	OutputPath string
}

func printHelp() {
	color.Set(color.Bold)
	color.Set(color.FgHiYellow)
	println("Action           Flags")

	color.Set(color.Reset)
	println("build [target]")
	println("                 -to --tokens            Prints lexed tokens.")
	println("                 -tr --tree              Draws abstract syntax tree.")
	println("                 -i  --instructions      Prints generated instructions.")
	println("                 -d  --dontOptimize      Compiler won't optimize byte code.")
	println("                 -s  --silent            Doesn't produce info messages when possible.")
	println("                 -n  --noLog             Doesn't produce any log messages, even if there are errors.")
	println("                 -l  --logLevel [LEVEL]  Sets logging level. Possible values are 0 to 5.")
	println("                 -o  --out               Sets output file path.")
	println("\nrun [target]")
	println("\nanalyze [target]")
	println("                 -to --tokens        Prints lexed tokens.")
	println("                 -tr --tree          Draws abstract syntax tree.")
	println("                 -d  --dontOptimize  Compiler won't optimize byte code.")

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

func printInstructions(instructions *[]VM.Instruction, constants []any, firstLine int) {
	line := firstLine
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

		// Print instruction number,
		if i < 10 {
			print(" ")
		}
		fmt.Printf("%d  ", i)

		// Print instruction name
		fmt.Printf("%s", VM.InstructionTypeToString[instruction.InstructionType])

		j := len(VM.InstructionTypeToString[instruction.InstructionType])
		for j < 16 {
			print(" ")
			j++
		}

		// Print arguments
		if len(instruction.InstructionValue) != 0 {
			fmt.Printf("%d", instruction.InstructionValue[0])

			if instruction.InstructionType == VM.IT_JumpBack {
				fmt.Printf(" (%d)", i-int(instruction.InstructionValue[0])+1)
			} else if instruction.InstructionType == VM.IT_Jump || instruction.InstructionType == VM.IT_JumpIfTrue {
				fmt.Printf(" (%d)", i+int(instruction.InstructionValue[0])+1)
			} else if instruction.InstructionType == VM.IT_PushScope || instruction.InstructionType == VM.IT_LoadConst || instruction.InstructionType == VM.IT_LoadConstToList {
				if reflect.TypeOf(constants[instruction.InstructionValue[0]]).Kind() == reflect.String {
					fmt.Printf("  (\"%v\")", constants[instruction.InstructionValue[0]])
				} else {
					fmt.Printf("  (%v)", constants[instruction.InstructionValue[0]])
				}
			} else if instruction.InstructionType == VM.IT_CallBuiltInFunc {
				fmt.Printf("  %v()", VM.BuiltInFuncToString[instruction.InstructionValue[0]])
			}
		}

		println()
	}
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
			case "--dontOptimize", "-d":
				configuration.Optimize = false
			case "--silent", "-s":
				logger.LoggingLevel = logger.LL_Error
			case "--noLog", "-n":
				logger.LoggingLevel = logger.LL_NoLog
			case "--logLevel", "-l":
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
			case "--dontOptimize", "-d":
				configuration.Optimize = false
			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action analyze.")
			}
		}
	}

	// Set output binary path
	if configuration.OutputPath == "" {
		configuration.OutputPath = configuration.TargetPath[:len(configuration.TargetPath)-5]
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

func compile(configuration *Configuration) {
	if !configuration.Optimize {
		logger.Warning("Code optimization disabled.")
	}

	startTime := time.Now()

	tree, p := analyze(configuration)

	// Generate code
	codeGenerator := codeGen.NewGenerator(tree, p.IntConstants, p.FloatConstants, p.StringConstants, configuration.Optimize)
	codeGenerator.Generate()

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

		println()
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
		virtualMachine := VM.NewVirutalMachine()
		virtualMachine.Execute(configuration.TargetPath)
		// Analyze target
	} else if configuration.Action == A_Analyze {
		logger.Info("🐱 Analyzing " + configuration.TargetPath)
		startTime := time.Now()

		analyze(configuration)

		logger.Success(fmt.Sprintf("😺 Analyze completed in %s.", time.Since(startTime)))
	}
}
