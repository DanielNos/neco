package main

import (
	"fmt"
	"reflect"

	"github.com/DanielNos/NeCo/lexer"
	VM "github.com/DanielNos/NeCo/virtualMachine"

	"github.com/fatih/color"
)

func printTokens(tokens []*lexer.Token) {
	fmt.Println()
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
	fmt.Println()
}

func printInstructions(instructions *[]VM.Instruction, constants []any, firstLine int) {
	line := firstLine
	justChanged := true

	for i, instruction := range *instructions {
		// Skip removed instruction
		if instruction.InstructionType == 255 {
			continue
		}

		// Display empty line instead of offset and record new line number
		if instruction.InstructionType == VM.IT_LineOffset {
			line += int(instruction.InstructionValue[0])
			justChanged = true

			fmt.Println()
			continue
		}

		// Print line number
		if justChanged {
			if line < 10 {
				fmt.Print(" ")
			}
			fmt.Printf("%d ", line)
			justChanged = false
		} else {
			fmt.Print("   ")
		}

		// Print instruction number,
		if i < 10 {
			fmt.Print(" ")
		}
		if i < 100 {
			fmt.Print(" ")
		}
		fmt.Printf("%d  ", i)

		// Print instruction name
		fmt.Printf("%s", VM.InstructionTypeToString[instruction.InstructionType])

		j := len(VM.InstructionTypeToString[instruction.InstructionType])
		for j < 16 {
			fmt.Print(" ")
			j++
		}

		// Print arguments
		if len(instruction.InstructionValue) != 0 {
			fmt.Printf("%d", instruction.InstructionValue[0])

			// Jump back instructions
			if instruction.InstructionType == VM.IT_JumpBack || instruction.InstructionType == VM.IT_JumpBackEx {
				fmt.Printf(" (%d)", i-int(instruction.InstructionValue[0])+1)

				// Jump forward instructions
			} else if VM.IsJumpForward(instruction.InstructionType) {
				fmt.Printf(" (%d)", i+int(instruction.InstructionValue[0])+1)

				// Instructions loading constants
			} else if instruction.InstructionType == VM.IT_PushScope || instruction.InstructionType == VM.IT_LoadConst || instruction.InstructionType == VM.IT_LoadConstToList {

				if reflect.TypeOf(constants[instruction.InstructionValue[0]]).Kind() == reflect.String {
					fmt.Printf("  (\"%v\")", constants[instruction.InstructionValue[0]])
				} else {
					fmt.Printf("  (%v)", constants[instruction.InstructionValue[0]])
				}

				// Instruction calling built-in functions
			} else if instruction.InstructionType == VM.IT_CallBuiltInFunc {
				fmt.Printf("  %v()", VM.BuiltInFuncToString[instruction.InstructionValue[0]])
			}
		}

		fmt.Println()
	}
}

func printConstants(stringsCount, intsCount, floatsCount int, constants []any) {
	// Calculate segments sizes
	stringsSize := 0
	for i := 0; i < stringsCount; i++ {
		stringsSize += len(constants[i].(string)) + 1
	}

	intsSize := intsCount * 8
	floatsSize := floatsCount * 8

	color.Yellow("Constants %d B\n", stringsSize+intsSize+floatsSize)

	// Print constants
	color.Set(color.FgHiWhite)
	index := 0

	fmt.Print("├─ ")
	color.HiYellow("Strings %d B\n", stringsSize)

	for index < stringsCount {
		if index == stringsCount-1 {
			fmt.Print("│  └─ ")
		} else {
			fmt.Print("│  ├─ ")
		}
		fmt.Printf("[%d] \"%v\"\n", index, constants[index])
		index++
	}
	fmt.Println("│")

	fmt.Print("├─ ")
	color.HiYellow("Integers %d B\n", intsSize)

	endOfInts := stringsCount + intsCount
	for index < endOfInts {
		if index == endOfInts-1 {
			fmt.Print("│  └─ ")
		} else {
			fmt.Print("│  ├─ ")
		}
		fmt.Printf("[%d] %v\n", index, constants[index])
		index++
	}
	fmt.Println("│")

	fmt.Print("└─ ")
	color.HiYellow("Floats %d B\n", floatsSize)

	for index < len(constants) {
		if index == len(constants)-1 {
			fmt.Print("   └─ ")
		} else {
			fmt.Print("   ├─ ")
		}
		fmt.Printf("[%d] %v\n", index, constants[index])
		index++
	}
}
