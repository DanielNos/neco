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

func printHelp() {
	color.Set(color.Bold)
	color.Set(color.FgHiYellow)
	println("Action           Flags")

	color.Set(color.Reset)
	println("build [target]")
	println("                 -to --tokens           Prints lexed tokens.")
	println("                 -tr --tree             Draws abstract syntax tree.")
	println("                 -i  --instructions     Prints generated instructions.")
	println("                 -d  --dontOptimize     Compiler won't optimize byte code.")
	println("                 -s  --silent           Doesn't produce info messages when possible.")
	println("                 -n  --noLog            Doesn't produce any log messages, even if there are errors.")
	println("                 -l  --logLevel [LEVEL] Sets logging level. Possible values are 0 to 5.")
	println("\nrun [target]")
	println("\nanalyze [target] -to --tokens  Prints lexed tokens.")
	println("                 -tr --tree    Draws abstract syntax tree.")
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

func processArguments() (string, string, []bool) {
	args := os.Args[1:]

	// No action
	if len(args) == 0 {
		logger.Fatal(errors.INVALID_FLAGS, "No action specified. Use neco help for more info.")
	}

	action := args[0]
	argumentsStart := 2

	// Collect target
	target := ""

	switch action {
	case "build", "run", "analyze":
		if len(args) == 1 {
			logger.Fatal(errors.INVALID_FLAGS, "No target specified.")
		}
		target = args[1]
	case "help", "--help", "-h":
		printHelp()
		os.Exit(0)
	default:
		if strings.HasSuffix(args[0], ".neco") {
			action = "build"
		} else {
			action = "run"
		}

		target = args[0]
		argumentsStart = 1
	}

	// Collect flags
	var flags []bool

	switch action {
	// Build flags
	case "build":
		flags = []bool{false, false, false, false}
		for i := argumentsStart; i < len(args); i++ {
			switch args[i] {
			case "--tokens", "-to":
				flags[0] = true
			case "--tree", "-tr":
				flags[1] = true
			case "--instructions", "-i":
				flags[2] = true
			case "--dontOptimize", "-d":
				flags[3] = true
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

			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+args[i]+"\" for action build.")
			}
		}
	// Run flags
	case "run":
		for _, flag := range args[argumentsStart:] {
			switch flag {
			default:
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action run.")
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
				logger.Fatal(errors.INVALID_FLAGS, "Invalid flag \""+flag+"\" for action analyze.")
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

func compile(path string, showTokens, showTree, printInstruction, optimize bool) {
	startTime := time.Now()

	tree, p := analyze(path, showTokens, showTree, true)

	// Generate code
	codeGenerator := codeGen.NewGenerator(tree, path[:len(path)-5], p.IntConstants, p.FloatConstants, p.StringConstants, optimize)
	codeGenerator.Generate()

	// Generation failed
	if codeGenerator.ErrorCount != 0 {
		logger.Fatal(errors.CODE_GENERATION, fmt.Sprintf("Failed code generation with %d error/s.", codeGenerator.ErrorCount))
	}

	logger.Info(fmt.Sprintf("Generated %d instructions.", len(codeGenerator.GlobalsInstructions)+len(codeGenerator.FunctionsInstructions)))
	logger.Success(fmt.Sprintf("😺 Compilation completed in %s.", time.Since(startTime)))

	codeWriter := codeGen.NewCodeWriter(codeGenerator)
	codeWriter.Write()

	// Print generated instructions
	if printInstruction {
		printInstructions(&codeGenerator.GlobalsInstructions, codeGenerator.Constants, int(codeGenerator.FirstLine))
		printInstructions(&codeGenerator.FunctionsInstructions, codeGenerator.Constants, int(codeGenerator.FirstLine))

		println()
	}
}

func main() {
	action, target, flags := processArguments()

	// Build target
	if action == "build" {
		logger.Info("🐱 Compiling " + target)
		compile(target, flags[0], flags[1], flags[2], !flags[3])
		// Run target
	} else if action == "run" {
		virtualMachine := VM.NewVirutalMachine()
		virtualMachine.Execute(target)
		// Analyze target
	} else if action == "analyze" {
		logger.Info("🐱 Analyzing " + target)
		startTime := time.Now()

		analyze(target, flags[0], flags[1], false)

		logger.Success(fmt.Sprintf("😺 Analyze completed in %s.", time.Since(startTime)))
	}
}
