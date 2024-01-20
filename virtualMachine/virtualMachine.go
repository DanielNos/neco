package virtualMachine

import (
	"neko/dataStructures"
	"neko/parser"
)

const (
	VERSION_MAJOR byte = 0
	VERSION_MINOR      = 1
	VERSION_PATCH      = 0
)

const (
	Reg_GenericA byte = iota
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
		// Load Constant
		case IT_LoadConstant:
			switch instruction.InstructionValue[1] {
			case Reg_GenericA:
				vm.Reg_GenericA = vm.Constants[instruction.InstructionValue[0]]

			case Reg_GenericB:
				vm.Reg_GenericB = vm.Constants[instruction.InstructionValue[0]]

			case Stack_Argument:
				vm.Stack_Argument[vm.Reg_ArgumentPointer] = vm.Constants[instruction.InstructionValue[0]]
				vm.Reg_ArgumentPointer++
			}

		// Push
		case IT_Push:
			var data interface{}
			switch instruction.InstructionValue[0] {
			case Reg_GenericA:
				data = vm.Reg_GenericA
			case Reg_GenericB:
				data = vm.Reg_GenericB
			}

			switch instruction.InstructionValue[1] {
			case Stack_Argument:
				vm.Stack_Argument[vm.Reg_ArgumentPointer] = data
				vm.Reg_ArgumentPointer++
			}

		// Swap generic registers
		case IT_SwapGeneric:
			vm.Reg_GenericA, vm.Reg_GenericB = vm.Reg_GenericB, vm.Reg_GenericA

		// Add operator
		case IT_IntAdd:

		// Call built-in function
		case IT_CallBuiltInFunction:
			vm.callBuiltInFunction(instruction.InstructionValue[0])

		// Declare variables
		case IT_DeclareBool:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Bool, nil}})

		case IT_DeclareInt:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Int, nil}})

		case IT_DeclareFloat:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_Float, nil}})

		case IT_DeclareString:
			vm.SymbolTables.Top.Value = append(vm.SymbolTables.Top.Value.([]Symbol), Symbol{ST_Variable, VariableSymbol{parser.DT_String, nil}})

		// Load variable
		case IT_LoadRegisterA:
			vm.Reg_GenericA = vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue

		// Store variable
		case IT_StoreRegisterA:
			vm.SymbolTables.Top.Value.([]Symbol)[instruction.InstructionValue[0]].symbolValue = vm.Reg_GenericA

		// Move line
		case IT_LineOffset:
			vm.Line += uint(instruction.InstructionValue[0])
		}
	}
}
