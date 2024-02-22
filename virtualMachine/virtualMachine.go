package virtualMachine

import (
	"bufio"
	"fmt"
	"math"
	data "neco/dataStructures"
	"neco/errors"
	"neco/logger"
	"os"
)

const (
	stack_arguments_SIZE    = 100
	STACK_RETURN_INDEX_SIZE = 1024
	stack_scopes_SIZE       = 100
	SYMBOL_MAP_SIZE         = 100
)

var InstructionToDataType = map[byte]data.DType{
	IT_DeclareBool:   data.DT_Bool,
	IT_DeclareInt:    data.DT_Int,
	IT_DeclareFloat:  data.DT_Float,
	IT_DeclareString: data.DT_String,
	IT_DeclareList:   data.DT_List,
}

type VirtualMachine struct {
	Constants []interface{}

	Instructions     []ExpandedInstruction
	instructionIndex int

	functions []int

	// Public registers and stack
	reg_operationA     interface{} // A) Operation A
	reg_operationB     interface{} // B) Operation B
	reg_operationStore interface{} // C) Operation Store
	reg_listA          interface{} // D) List A
	reg_funcReturnA    interface{} // F) Function Return
	reg_funcReturnB    interface{} // G) Function error return

	reg_argumentPointer int
	stack_arguments     []interface{}

	// Private stacks
	reg_returnIndex     int
	stack_returnIndexes []int

	reg_scopeIndex int
	stack_scopes   []string

	reg_symbolIndex    int
	stack_symbolTables *data.Stack

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
		stack_symbolTables: data.NewStack(),

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

		// Load list element
		case IT_LoadListAtOpAToOpA:
			list := vm.findSymbol().symbolValue.(*VariableSymbol).value

			if vm.reg_operationA.(int64) > int64(len(list.([]interface{}))-1) {
				vm.traceLine()
				logger.Fatal(errors.INDEX_OUT_OF_RANGE, fmt.Sprintf("line %d: Index out of range. Index: %d, list size: %d.", vm.firstLine, vm.reg_operationA.(int64), len(list.([]interface{}))))
			}
			vm.reg_operationA = list.([]interface{})[vm.reg_operationA.(int64)]

		case IT_LoadListOpBToOpB:
			list := vm.findSymbol().symbolValue.(*VariableSymbol).value

			if vm.reg_operationB.(int64) > int64(len(list.([]interface{}))-1) {
				vm.traceLine()
				logger.Fatal(errors.INDEX_OUT_OF_RANGE, fmt.Sprintf("line %d: Index out of range. Index: %d, list size: %d.", vm.firstLine, vm.reg_operationB.(int64), len(list.([]interface{}))))
			}
			vm.reg_operationB = list.([]interface{})[vm.reg_operationB.(int64)]

		// Store register to a variable
		case IT_StoreOpA:
			vm.findSymbol().symbolValue.(*VariableSymbol).value = vm.reg_operationA

		case IT_StoreOpB:
			vm.findSymbol().symbolValue.(*VariableSymbol).value = vm.reg_operationB

		// List operations
		case IT_SetListAtAToB:
			vm.findSymbol().symbolValue.(*VariableSymbol).value.([]interface{})[vm.reg_operationA.(int64)] = vm.reg_operationB

		// Load constant to register
		case IT_LoadConstRegA:
			vm.reg_operationA = vm.Constants[instruction.InstructionValue[0]]

		case IT_LoadConstRegB:
			vm.reg_operationB = vm.Constants[instruction.InstructionValue[0]]

		case IT_LoadConstArgStack:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.Constants[instruction.InstructionValue[0]]
			vm.reg_argumentPointer++

		// Load variable to a register
		case IT_LoadRegA:
			vm.reg_operationA = vm.findSymbol().symbolValue.(*VariableSymbol).value

		case IT_LoadRegB:
			vm.reg_operationB = vm.findSymbol().symbolValue.(*VariableSymbol).value

		case IT_LoadArgStack:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.findSymbol().symbolValue.(*VariableSymbol).value
			vm.reg_argumentPointer++

		// Enter scope
		case IT_PushScope:
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
				vm.traceLine()
				logger.Fatal(errors.STACK_OVERFLOW, fmt.Sprintf("line %d: Function return adress stack overflow.", vm.firstLine))
			}

			// Jump to function
			vm.instructionIndex = vm.functions[instruction.InstructionValue[0]] - 1
			continue

		// Declare variables
		case IT_DeclareBool:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Bool, nil}, nil}})

		case IT_DeclareInt:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Int, nil}, nil}})

		case IT_DeclareFloat:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Float, nil}, nil}})

		case IT_DeclareString:
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_String, nil}, nil}})

		case IT_DeclareList:
			vm.instructionIndex++
			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_List, InstructionToDataType[vm.Instructions[vm.instructionIndex+1].InstructionType]}, []interface{}{}}})

		// NO ARGUMENT INSTRUCTIONS -------------------------------------------------------------------------

		// Swap generic registers
		case IT_SwapOperation:
			vm.reg_operationA, vm.reg_operationB = vm.reg_operationB, vm.reg_operationA

		// Copy registers to registers
		case IT_CopyOpAToOpStore:
			vm.reg_operationStore = vm.reg_operationA

		case IT_CopyOpBToOpStore:
			vm.reg_operationStore = vm.reg_operationB

		case IT_CopyOpStoreToOpA:
			vm.reg_operationA = vm.reg_operationStore

		case IT_CopyOpStoreToOpB:
			vm.reg_operationB = vm.reg_operationStore

		case IT_CopyOpAToReturn:
			vm.reg_funcReturnA = vm.reg_operationA

		case IT_CopyReturnToOpA:
			vm.reg_operationA = vm.reg_funcReturnA

		case IT_CopyReturnToOpB:
			vm.reg_operationB = vm.reg_funcReturnA

		case IT_CopyOpAToListA:
			vm.reg_listA = vm.reg_operationA

		case IT_CopyListAToOpA:
			vm.reg_operationA = vm.reg_listA

		case IT_CopyListAToOpB:
			vm.reg_operationB = vm.reg_listA

		// Push register to stack
		case IT_PushOpAToArg:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.reg_operationA
			vm.reg_argumentPointer++

		case IT_PushOpBToArg:
			vm.stack_arguments[vm.reg_argumentPointer] = vm.reg_operationB
			vm.reg_argumentPointer++

		// Integer operations
		case IT_IntAdd:
			vm.reg_operationA = vm.reg_operationA.(int64) + vm.reg_operationB.(int64)

		case IT_IntSubtract:
			vm.reg_operationA = vm.reg_operationA.(int64) - vm.reg_operationB.(int64)

		case IT_IntMultiply:
			vm.reg_operationA = vm.reg_operationA.(int64) * vm.reg_operationB.(int64)

		case IT_IntDivide:
			vm.reg_operationA = vm.reg_operationA.(int64) / vm.reg_operationB.(int64)

		case IT_IntPower:
			vm.reg_operationA = PowerInt64(vm.reg_operationA.(int64), vm.reg_operationB.(int64))

		case IT_IntModulo:
			vm.reg_operationA = vm.reg_operationA.(int64) % vm.reg_operationB.(int64)

		// Float operations
		case IT_FloatAdd:
			vm.reg_operationA = vm.reg_operationA.(float64) + vm.reg_operationB.(float64)

		case IT_FloatSubtract:
			vm.reg_operationA = vm.reg_operationA.(float64) - vm.reg_operationB.(float64)

		case IT_FloatMultiply:
			vm.reg_operationA = vm.reg_operationA.(float64) * vm.reg_operationB.(float64)

		case IT_FloatDivide:
			vm.reg_operationA = vm.reg_operationA.(float64) / vm.reg_operationB.(float64)

		case IT_FloatPower:
			vm.reg_operationA = math.Pow(vm.reg_operationA.(float64), vm.reg_operationB.(float64))

		case IT_FloatModulo:
			vm.reg_operationA = math.Mod(vm.reg_operationA.(float64), vm.reg_operationB.(float64))

		// String operations
		case IT_StringConcat:
			vm.reg_operationA = fmt.Sprintf("%s%s", vm.reg_operationA, vm.reg_operationB)

		// Return from a function
		case IT_Return:
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

			vm.reg_returnIndex--
			vm.instructionIndex = vm.stack_returnIndexes[vm.reg_returnIndex]
			continue

		// Comparison instructions
		case IT_Equal:
			vm.reg_operationA = vm.reg_operationA == vm.reg_operationB

		case IT_IntLower:
			vm.reg_operationA = vm.reg_operationA.(int64) < vm.reg_operationB.(int64)

		case IT_FloatLower:
			vm.reg_operationA = vm.reg_operationA.(float64) < vm.reg_operationB.(float64)

		case IT_IntGreater:
			vm.reg_operationA = vm.reg_operationA.(int64) > vm.reg_operationB.(int64)

		case IT_FloatGreater:
			vm.reg_operationA = vm.reg_operationA.(float64) > vm.reg_operationB.(float64)

		case IT_IntLowerEqual:
			vm.reg_operationA = vm.reg_operationA.(int64) <= vm.reg_operationB.(int64)

		case IT_FloatLowerEqual:
			vm.reg_operationA = vm.reg_operationA.(float64) <= vm.reg_operationB.(float64)

		case IT_IntGreaterEqual:
			vm.reg_operationA = vm.reg_operationA.(int64) >= vm.reg_operationB.(int64)

		case IT_FloatGreaterEqual:
			vm.reg_operationA = vm.reg_operationA.(float64) >= vm.reg_operationB.(float64)

		case IT_Not:
			vm.reg_operationA = !vm.reg_operationA.(bool)

		// Jumps
		case IT_JumpBack:
			vm.instructionIndex -= instruction.InstructionValue[0]

		case IT_Jump:
			vm.instructionIndex += instruction.InstructionValue[0]

		case IT_JumpIfTrue:
			if vm.reg_operationA.(bool) {
				vm.instructionIndex += instruction.InstructionValue[0]
			}

		// Put bools in registers
		case IT_SetRegATrue:
			vm.reg_operationA = true

		case IT_SetRegAFalse:
			vm.reg_operationA = false

		case IT_SetRegBTrue:
			vm.reg_operationA = true

		case IT_SetRegBFalse:
			vm.reg_operationA = false

		// Scopes
		case IT_PushScopeUnnamed:
			vm.stack_scopes[vm.reg_scopeIndex] = ""
			vm.reg_scopeIndex++

			vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

		case IT_PopScope:
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

		// List operations
		case IT_CreateListInListA:
			vm.reg_listA = []interface{}{}

		case IT_AppendOpAToListA:
			vm.reg_listA = append(vm.reg_listA.([]interface{}), vm.reg_operationA)

		case IT_ListConcat:
			vm.reg_operationA = append(vm.reg_operationA.([]interface{}), vm.reg_operationB.([]interface{})...)

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

func (vm *VirtualMachine) findSymbol() *Symbol {
	// Find variable
	symbolTable := vm.stack_symbolTables.Top
	value := symbolTable.Value.(*SymbolMap).Get(vm.Instructions[vm.instructionIndex].InstructionValue[0])

	for value == nil {
		symbolTable = symbolTable.Previous
		value = symbolTable.Value.(*SymbolMap).Get(vm.Instructions[vm.instructionIndex].InstructionValue[0])
	}

	// Couldn't find variable
	if value == nil {
		vm.traceLine()
		logger.Fatal(errors.UNDECLARED_VARIABLE, fmt.Sprintf("line %d: Undeclared variable %d.", vm.firstLine, vm.Instructions[vm.instructionIndex].InstructionValue))
	}

	return value
}
