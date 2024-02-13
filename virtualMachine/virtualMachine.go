package virtualMachine

import (
	"bufio"
	"fmt"
	"math"
	"neco/dataStructures"
	"neco/errors"
	"neco/logger"
	"neco/necoMath"
	"neco/parser"
	"os"
)

const (
	stack_arguments_SIZE    = 100
	STACK_RETURN_INDEX_SIZE = 1024
	stack_scopes_SIZE       = 100
	SYMBOL_MAP_SIZE         = 100
)

type VirtualMachine struct {
	Constants []interface{}

	Instructions     []ExpandedInstruction
	instructionIndex int

	functions []int

	// Public registers and stack
	reg_genericA interface{}
	reg_genericB interface{}
	reg_genericC interface{}
	reg_genericD interface{}

	reg_argumentPointer int
	stack_arguments     []interface{}

	// Private stacks
	reg_returnIndex     int
	stack_returnIndexes []int

	reg_scopeIndex int
	stack_scopes   []string

	reg_symbolIndex    int
	stack_symbolTables *dataStructures.Stack

	reader    *bufio.Reader
	firstLine int
}

func NewVirutalMachine() *VirtualMachine {
	virtualMachine := &VirtualMachine{
		instructionIndex: 0,

		stack_arguments: make([]interface{}, stack_arguments_SIZE),

		reg_returnIndex:     0,
		stack_returnIndexes: make([]int, STACK_RETURN_INDEX_SIZE),

		reg_scopeIndex: 0,
		stack_scopes:   make([]string, stack_scopes_SIZE),

		reg_symbolIndex:    0,
		stack_symbolTables: dataStructures.NewStack(),

		reader: bufio.NewReader(os.Stdin),
	}

	virtualMachine.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	return virtualMachine
}

func (vm *VirtualMachine) Execute(filePath string) {
	// Read instructions
	reader := NewInstructionReader(filePath, vm)
	reader.Read()

	// Enter root scope
	vm.stack_returnIndexes[vm.reg_returnIndex] = len(vm.Instructions)
	vm.reg_returnIndex++

	vm.stack_scopes[vm.reg_scopeIndex] = filePath
	vm.reg_scopeIndex++

	vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	for vm.instructionIndex < len(vm.Instructions) {
		instruction := vm.Instructions[vm.instructionIndex]

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
			// Find variable
			symbolTable := vm.stack_symbolTables.Top
			value := symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])

			for value == nil {
				symbolTable = symbolTable.Previous
				value = symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])
			}

			// Couldn't find variable
			if value == nil {
				vm.traceLine()
				logger.Fatal(errors.UNDECLARED_VARIABLE, fmt.Sprintf("Undecalred variable %d.", instruction.InstructionValue))
			}

			// Store register B in symbol
			value.symbolValue = vm.reg_genericA

		case IT_StoreRegB:
			// Find variable
			symbolTable := vm.stack_symbolTables.Top
			value := symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])

			for value == nil {
				symbolTable = symbolTable.Previous
				value = symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])
			}

			// Couldn't find variable
			if value == nil {
				vm.traceLine()
				logger.Fatal(errors.UNDECLARED_VARIABLE, fmt.Sprintf("Undecalred variable %d.", instruction.InstructionValue))
			}

			// Store register B in symbol
			value.symbolValue = vm.reg_genericB

		// Load constant to register
		case IT_LoadConstRegA:
			vm.reg_genericA = vm.Constants[instruction.InstructionValue[0]]

		case IT_LoadConstRegB:
			vm.reg_genericB = vm.Constants[instruction.InstructionValue[0]]

		// Load variable to a register
		case IT_LoadRegA:
			// Find variable
			symbolTable := vm.stack_symbolTables.Top
			value := symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])

			for value == nil {
				symbolTable = symbolTable.Previous
				value = symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])
			}

			// Couldn't find variable
			if value == nil {
				vm.traceLine()
				logger.Fatal(errors.UNDECLARED_VARIABLE, fmt.Sprintf("Undecalred variable %d.", instruction.InstructionValue))
			}

			vm.reg_genericA = value.symbolValue

		case IT_LoadRegB:
			// Find variable
			symbolTable := vm.stack_symbolTables.Top
			value := symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])

			for value == nil {
				symbolTable = symbolTable.Previous
				value = symbolTable.Value.(*SymbolMap).Get(instruction.InstructionValue[0])
			}

			// Couldn't find variable
			if value == nil {
				vm.traceLine()
				logger.Fatal(errors.UNDECLARED_VARIABLE, fmt.Sprintf("Undecalred variable %d.", instruction.InstructionValue))
			}

			vm.reg_genericB = value.symbolValue

		// Enter scope
		case IT_PushScope:
			fmt.Printf("ENTERING %s\n", vm.Constants[instruction.InstructionValue[0]].(string))
			vm.stack_scopes[vm.reg_scopeIndex] = vm.Constants[instruction.InstructionValue[0]].(string)
			vm.reg_scopeIndex++

			vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

		// Call a function
		case IT_Call:
			// Push return adress to stack
			vm.stack_returnIndexes[vm.reg_returnIndex] = vm.instructionIndex + 1
			vm.reg_returnIndex++

			// Return adress stack overflow
			if vm.reg_returnIndex == STACK_RETURN_INDEX_SIZE {
				logger.Fatal(errors.STACK_OVERFLOW, "Function return adress stack overflow.")
			}

			// Jump to function
			vm.instructionIndex = vm.functions[instruction.InstructionValue[0]] - 1
			continue

		// NO ARGUMENT INSTRUCTIONS -------------------------------------------------------------------------

		// Swap generic registers
		case IT_SwapAB:
			vm.reg_genericA, vm.reg_genericB = vm.reg_genericB, vm.reg_genericA

		// Copy registers to registers
		case IT_CopyRegAToC:
			vm.reg_genericC = vm.reg_genericA

		case IT_CopyRegBToC:
			vm.reg_genericC = vm.reg_genericB

		case IT_CopyRegCToA:
			vm.reg_genericA = vm.reg_genericC

		case IT_CopyRegCToB:
			vm.reg_genericB = vm.reg_genericC

		case IT_CopyRegAToD:
			vm.reg_genericD = vm.reg_genericA

		case IT_CopyRegDToA:
			vm.reg_genericA = vm.reg_genericD

		case IT_CopyRegDToB:
			vm.reg_genericB = vm.reg_genericD

		// Push register to stack
		case IT_PushRegAToArgStack:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.reg_genericA
			vm.reg_argumentPointer++

		case IT_PushRegBToArgStack:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.reg_genericB
			vm.reg_argumentPointer++

		// Integer operations
		case IT_IntAdd:
			vm.reg_genericA = vm.reg_genericA.(int64) + vm.reg_genericB.(int64)

		case IT_IntSubtract:
			vm.reg_genericA = vm.reg_genericA.(int64) - vm.reg_genericB.(int64)

		case IT_IntMultiply:
			vm.reg_genericA = vm.reg_genericA.(int64) * vm.reg_genericB.(int64)

		case IT_IntDivide:
			vm.reg_genericA = vm.reg_genericA.(int64) / vm.reg_genericB.(int64)

		case IT_IntPower:
			vm.reg_genericA = necoMath.PowerInt64(vm.reg_genericA.(int64), vm.reg_genericB.(int64))

		case IT_IntModulo:
			vm.reg_genericA = vm.reg_genericA.(int64) % vm.reg_genericB.(int64)

		// Float operations
		case IT_FloatAdd:
			vm.reg_genericA = vm.reg_genericA.(float64) + vm.reg_genericB.(float64)

		case IT_FloatSubtract:
			vm.reg_genericA = vm.reg_genericA.(float64) - vm.reg_genericB.(float64)

		case IT_FloatMultiply:
			vm.reg_genericA = vm.reg_genericA.(float64) * vm.reg_genericB.(float64)

		case IT_FloatDivide:
			vm.reg_genericA = vm.reg_genericA.(float64) / vm.reg_genericB.(float64)

		case IT_FloatPower:
			vm.reg_genericA = math.Pow(vm.reg_genericA.(float64), vm.reg_genericB.(float64))

		case IT_FloatModulo:
			vm.reg_genericA = math.Mod(vm.reg_genericA.(float64), vm.reg_genericB.(float64))

		// String operations
		case IT_StringConcat:
			vm.reg_genericA = fmt.Sprintf("%s%s", vm.reg_genericA, vm.reg_genericB)

		// Declare variables
		case IT_DeclareBool:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, VariableSymbol{parser.DT_Bool, nil}})

		case IT_DeclareInt:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, VariableSymbol{parser.DT_Int, nil}})

		case IT_DeclareFloat:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, VariableSymbol{parser.DT_Float, nil}})

		case IT_DeclareString:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, VariableSymbol{parser.DT_String, nil}})

		// Return from a function
		case IT_Return:
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

			vm.reg_returnIndex--
			vm.instructionIndex = vm.stack_returnIndexes[vm.reg_returnIndex]
			continue

		// Comparison instructions
		case IT_Equal:
			vm.reg_genericA = vm.reg_genericA == vm.reg_genericB

		case IT_LowerInt:
			vm.reg_genericA = vm.reg_genericA.(int64) < vm.reg_genericB.(int64)

		case IT_LowerFloat:
			vm.reg_genericA = vm.reg_genericA.(float64) < vm.reg_genericB.(float64)

		case IT_GreaterInt:
			vm.reg_genericA = vm.reg_genericA.(int64) > vm.reg_genericB.(int64)

		case IT_GreaterFloat:
			vm.reg_genericA = vm.reg_genericA.(float64) > vm.reg_genericB.(float64)

		case IT_LowerEqualInt:
			vm.reg_genericA = vm.reg_genericA.(int64) <= vm.reg_genericB.(int64)

		case IT_LowerEqualFloat:
			vm.reg_genericA = vm.reg_genericA.(float64) <= vm.reg_genericB.(float64)

		case IT_GreaterEqualInt:
			vm.reg_genericA = vm.reg_genericA.(int64) >= vm.reg_genericB.(int64)

		case IT_GreaterEqualFloat:
			vm.reg_genericA = vm.reg_genericA.(float64) >= vm.reg_genericB.(float64)

		case IT_Not:
			vm.reg_genericA = !vm.reg_genericA.(bool)

			// Jumps
		case IT_JumpBack:
			vm.instructionIndex -= instruction.InstructionValue[0]

		case IT_Jump:
			vm.instructionIndex += instruction.InstructionValue[0]

		case IT_JumpIfTrue:
			if vm.reg_genericA.(bool) {
				vm.instructionIndex += instruction.InstructionValue[0]
			}

		// Put bools in registers
		case IT_SetRegATrue:
			vm.reg_genericA = true

		case IT_SetRegAFalse:
			vm.reg_genericA = false

		case IT_SetRegBTrue:
			vm.reg_genericA = true

		case IT_SetRegBFalse:
			vm.reg_genericA = false

		// Scopes
		case IT_PushScopeUnnamed:
			println("ENTERING UNNAMED")
			vm.stack_scopes[vm.reg_scopeIndex] = ""
			vm.reg_scopeIndex++

			vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

		case IT_PopScope:
			fmt.Printf("POPPING %s\n", vm.stack_scopes[vm.reg_scopeIndex])
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

		// Ignore line offsets
		case IT_LineOffset:

		// Unknown instruction
		default:
			vm.traceLine()
			logger.Fatal(errors.UNKNOWN_INSTRUCTION, fmt.Sprintf("line %d: Unknown instruction type: %d.", vm.firstLine, instruction.InstructionType))
		}

		vm.instructionIndex++
	}
}

func (vm *VirtualMachine) traceLine() {
	for i := 0; i < vm.instructionIndex; i++ {
		if vm.Instructions[i].InstructionType == IT_LineOffset {
			vm.firstLine += vm.Instructions[i].InstructionValue[0]
		}
	}
}
