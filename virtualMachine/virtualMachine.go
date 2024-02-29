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
	STACK_SIZE              = 1024
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

	stack *Stack

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

		stack: NewStack(STACK_SIZE),

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

	step := false

	for vm.instructionIndex < len(vm.Instructions) {
		instruction := vm.Instructions[vm.instructionIndex]

		switch instruction.InstructionType {

		// 1 ARGUMENT INSTRUCTIONS --------------------------------------------------------------------------

		// Jumps
		case IT_Jump:
			vm.instructionIndex += instruction.InstructionValue[0]

		case IT_JumpIfTrue:
			if vm.stack.Pop().(bool) {
				vm.instructionIndex += instruction.InstructionValue[0]
			}

		case IT_JumpBack:
			vm.instructionIndex -= instruction.InstructionValue[0]

		// Call functions
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

		case IT_CallBuiltInFunc:
			vm.callBuiltInFunction(instruction.InstructionValue[0])

		case IT_PushScope:
			vm.stack_scopes[vm.reg_scopeIndex] = vm.Constants[instruction.InstructionValue[0]].(string)
			vm.reg_scopeIndex++

			vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

		// Halt
		case IT_Halt:
			os.Exit(int(instruction.InstructionValue[0]))

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

			dataType := data.DataType{data.DT_List, nil}
			endType := &dataType.SubType

			for vm.Instructions[vm.instructionIndex].InstructionType == IT_DeclareList {
				dataType = data.DataType{data.DT_List, dataType}
				vm.instructionIndex++
			}

			*endType = InstructionToDataType[vm.Instructions[vm.instructionIndex].InstructionType]

			vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{dataType, []interface{}{}}})

		// Set and load list at index
		case IT_SetListAtPrevToCurr:
			vm.findSymbol().symbolValue.(*VariableSymbol).value.([]interface{})[vm.stack.Pop().(int64)] = vm.stack.Pop()

		// Load and store
		case IT_LoadConst:
			vm.stack.Push(vm.Constants[instruction.InstructionValue[0]])

		case IT_LoadConstToList:
			(*vm.stack.Top()) = append((*vm.stack.Top()).([]interface{}), vm.Constants[instruction.InstructionValue[0]])

		case IT_Load:
			vm.stack.Push(vm.findSymbol().symbolValue.(*VariableSymbol).value)

		case IT_Store:
			vm.findSymbol().symbolValue.(*VariableSymbol).value = vm.stack.Pop()

		// NO ARGUMENT INSTRUCTIONS -------------------------------------------------------------------------

		// Integer operations
		case IT_IntAdd:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) + vm.stack.items[vm.stack.size].(int64)

		case IT_IntSubtract:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) - vm.stack.items[vm.stack.size].(int64)

		case IT_IntMultiply:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) * vm.stack.items[vm.stack.size].(int64)

		case IT_IntDivide:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) / vm.stack.items[vm.stack.size].(int64)

		case IT_IntPower:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = PowerInt64(vm.stack.items[vm.stack.size-1].(int64), vm.stack.items[vm.stack.size].(int64))

		case IT_IntModulo:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) % vm.stack.items[vm.stack.size].(int64)

		// Float operations
		case IT_FloatAdd:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) + vm.stack.items[vm.stack.size].(float64)

		case IT_FloatSubtract:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) - vm.stack.items[vm.stack.size].(float64)

		case IT_FloatMultiply:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) * vm.stack.items[vm.stack.size].(float64)

		case IT_FloatDivide:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) / vm.stack.items[vm.stack.size].(float64)

		case IT_FloatPower:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = math.Pow(vm.stack.items[vm.stack.size-1].(float64), vm.stack.items[vm.stack.size].(float64))

		case IT_FloatModulo:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = math.Mod(vm.stack.items[vm.stack.size-1].(float64), vm.stack.items[vm.stack.size].(float64))

		// Concatenations
		case IT_StringConcat:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = fmt.Sprintf("%s%s", vm.stack.items[vm.stack.size-1].(string), vm.stack.items[vm.stack.size].(string))

		case IT_ListConcat:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = append(vm.stack.items[vm.stack.size-1].([]interface{}), vm.stack.items[vm.stack.size].([]interface{})...)

		// Return from a function
		case IT_Return:
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

			vm.reg_returnIndex--
			vm.instructionIndex = vm.stack_returnIndexes[vm.reg_returnIndex]
			continue

		// Comparison instructions
		case IT_Equal:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1] == vm.stack.items[vm.stack.size]

		case IT_IntLower:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) < vm.stack.items[vm.stack.size].(int64)

		case IT_FloatLower:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) < vm.stack.items[vm.stack.size].(float64)

		case IT_IntGreater:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) > vm.stack.items[vm.stack.size].(int64)

		case IT_FloatGreater:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) > vm.stack.items[vm.stack.size].(float64)

		case IT_IntLowerEqual:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) <= vm.stack.items[vm.stack.size].(int64)

		case IT_FloatLowerEqual:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) <= vm.stack.items[vm.stack.size].(float64)

		case IT_IntGreaterEqual:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(int64) >= vm.stack.items[vm.stack.size].(int64)

		case IT_FloatGreaterEqual:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(float64) >= vm.stack.items[vm.stack.size].(float64)

		case IT_Not:
			vm.stack.Push(!vm.stack.Pop().(bool))

		// Push boolens
		case IT_PushTrue:
			vm.stack.Push(true)

		case IT_PushFalse:
			vm.stack.Push(false)

		// Scopes
		case IT_PushScopeUnnamed:
			vm.stack_scopes[vm.reg_scopeIndex] = ""
			vm.reg_scopeIndex++

			vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

		case IT_PopScope:
			vm.stack_symbolTables.Pop()
			vm.reg_scopeIndex--

		// List operations
		case IT_CreateList:
			vm.stack.Push([]interface{}{})

		case IT_AppendToList:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = append(vm.stack.items[vm.stack.size-1].([]interface{}), vm.stack.items[vm.stack.size])

		case IT_IndexList:
			vm.stack.size--
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].([]interface{})[vm.stack.items[vm.stack.size].(int64)]

		// Ignore line offsets
		case IT_LineOffset:

		// Unknown instruction
		default:
			vm.traceLine()
			logger.Fatal(errors.UNKNOWN_INSTRUCTION, fmt.Sprintf("line %d: Unknown instruction type: %d.", vm.firstLine, instruction.InstructionType))
		}

		if step {
			fmt.Printf("Instrcution: %s %v\n", InstructionTypeToString[instruction.InstructionType], instruction.InstructionValue)
			fmt.Printf("Stack: %v\n", vm.stack.items[:vm.stack.size])
			print("Scope: {")
			fmt.Printf("%s", vm.stack_scopes[0])
			for _, scope := range vm.stack_scopes[1:vm.reg_scopeIndex] {
				fmt.Printf(", %s", scope)
			}
			println("}")
			fmt.Scanln()
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

	for value == nil && symbolTable.Previous != nil {
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
