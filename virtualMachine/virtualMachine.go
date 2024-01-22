package virtualMachine

import (
	"fmt"
	"neco/dataStructures"
	"neco/errors"
	"neco/logger"
	"neco/parser"
	"os"
)

const (
	VERSION_MAJOR byte = 0
	VERSION_MINOR      = 1
	VERSION_PATCH      = 0
)

type RegisterOrStack byte

const (
	Reg_GenericA RegisterOrStack = iota
	Reg_GenericB
	Stack_Argument
)

const STACK_ARGUMENT_SIZE = 100

type VirtualMachine struct {
	Constants    []interface{}
	Instructions []Instruction

	Reg_GenericA interface{}
	Reg_GenericB interface{}

	Reg_ArgumentPointer int
	Stack_Argument      []interface{}

	SymbolTables *dataStructures.Stack

	Line uint
}

func NewVirutalMachine() *VirtualMachine {
	virtualMachine := &VirtualMachine{Stack_Argument: make([]interface{}, STACK_ARGUMENT_SIZE), SymbolTables: dataStructures.NewStack()}

	virtualMachine.SymbolTables.Push([]Symbol{})

	return virtualMachine
}

func (vm *VirtualMachine) Execute(filePath string) {
	reader := NewInstructionReader(filePath, vm)
	reader.Read()

	for _, instruction := range vm.Instructions {
		switch instruction.InstructionType {

		// 1 ARGUMENT INSTRUCTIONS --------------------------------------------------------------------------

		// Call built-in function
		case IT_CallBuiltInFunc:
			vm.callBuiltInFunction(instruction.InstructionValue[0])

		// Halt
		case IT_Halt:
			os.Exit(int(instruction.InstructionValue[0]))

		// Store register to a variable
		case IT_StoreRegA:
			vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue = vm.Reg_GenericA

		case IT_StoreRegB:
			vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue = vm.Reg_GenericB

		// Load constant to register
		case IT_LoadConstRegA:
			vm.Reg_GenericA = vm.Constants[instruction.InstructionValue[0]]

		case IT_LoadConstRegB:
			vm.Reg_GenericB = vm.Constants[instruction.InstructionValue[0]]

		// Load variable to a register
		case IT_LoadRegA:
			vm.Reg_GenericA = vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue

		case IT_LoadRegB:
			vm.Reg_GenericB = vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue

		// NO ARGUMENT INSTRUCTIONS -------------------------------------------------------------------------

		// Swap generic registers
		case IT_SwapGeneric:
			vm.Reg_GenericA, vm.Reg_GenericB = vm.Reg_GenericB, vm.Reg_GenericA

		// Push register to stack
		case IT_PushRegAToArgStack:
			vm.Stack_Argument[vm.Reg_ArgumentPointer] = vm.Reg_GenericA
			vm.Reg_ArgumentPointer++

		case IT_PushRegBToArgStack:
			vm.Stack_Argument[vm.Reg_ArgumentPointer] = vm.Reg_GenericB
			vm.Reg_ArgumentPointer++

		// Integer operations
		case IT_IntAdd:
			vm.Reg_GenericA = vm.Reg_GenericA.(int64) + vm.Reg_GenericB.(int64)

		case IT_IntSubtract:
			vm.Reg_GenericA = vm.Reg_GenericA.(int64) - vm.Reg_GenericB.(int64)

		case IT_IntMultiply:
			vm.Reg_GenericA = vm.Reg_GenericA.(int64) * vm.Reg_GenericB.(int64)

		case IT_IntDivide:
			vm.Reg_GenericA = vm.Reg_GenericA.(int64) / vm.Reg_GenericB.(int64)

		case IT_IntPower:

			// Can't be done with negative exponent
			if vm.Reg_GenericB.(int64) < 0 {
				vm.Reg_GenericA = int64(0)
			} else {
				fmt.Printf("%d ^ %d\n", vm.Reg_GenericA.(int64), vm.Reg_GenericB.(int64))
				vm.Reg_GenericA = intPow(vm.Reg_GenericA.(int64), vm.Reg_GenericB.(int64))
			}

		case IT_IntModulo:
			vm.Reg_GenericA = vm.Reg_GenericA.(int64) % vm.Reg_GenericB.(int64)

		// String operations
		case IT_StringConcat:
			vm.Reg_GenericA = fmt.Sprintf("%s%s", vm.Reg_GenericA, vm.Reg_GenericB)

		// Declare variables
		case IT_DeclareBool:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Bool, nil}})

		case IT_DeclareInt:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Int, nil}})

		case IT_DeclareFloat:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Float, nil}})

		case IT_DeclareString:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_String, nil}})

		// Move line
		case IT_LineOffset:
			vm.Line += uint(instruction.InstructionValue[0])

		// Unknown instruction
		default:
			logger.Fatal(errors.ERROR_UNKNOWN_INSTRUCTION, fmt.Sprintf("line %d: Unknown instruction type: %d.", vm.Line, instruction.InstructionType))
		}
	}
}

func intPow(x, n int64) int64 {
	if n == 0 {
		return 1
	}

	if n == 1 {
		return x
	}

	y := intPow(x, n/2)
	if n%2 == 0 {
		return y * y
	}

	return x * y * y
}
