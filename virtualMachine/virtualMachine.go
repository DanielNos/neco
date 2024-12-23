package virtualMachine

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"

	data "github.com/DanielNos/neco/dataStructures"
)

const (
	STACK_SIZE              = 1024
	STACK_RETURN_INDEX_SIZE = 1024
	STACK_SCOPES_SIZE       = 256
	SYMBOL_MAP_SIZE         = 100
)

var InstructionToDataType = map[byte]data.PrimitiveType{
	IT_DeclareBool:   data.DT_Bool,
	IT_DeclareInt:    data.DT_Int,
	IT_DeclareFloat:  data.DT_Float,
	IT_DeclareString: data.DT_String,
	IT_DeclareList:   data.DT_List,
}

type VirtualMachine struct {
	Constants []any

	GlobalsInstructions   []ExpandedInstruction
	FunctionsInstructions []ExpandedInstruction

	instructions     *[]ExpandedInstruction
	instructionIndex int

	functions []int

	stack *Stack[any]

	reg_returnIndex     int
	stack_returnIndexes []int

	reg_scopeIndex int
	stack_scopes   []string

	reg_symbolIndex    int
	stack_symbolTables *data.Stack

	filePath string
	reader   *bufio.Reader
}

func NewVirtualMachine(filePath string) *VirtualMachine {
	virtualMachine := &VirtualMachine{
		instructionIndex: 0,

		stack: NewStack[any](STACK_SIZE),

		reg_returnIndex:     0,
		stack_returnIndexes: make([]int, STACK_RETURN_INDEX_SIZE),

		reg_scopeIndex: 0,
		stack_scopes:   make([]string, STACK_SCOPES_SIZE),

		reg_symbolIndex:    0,
		stack_symbolTables: data.NewStack(),

		filePath: filePath,
		reader:   bufio.NewReader(os.Stdin),
	}

	virtualMachine.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	return virtualMachine
}

var (
	currentObject object
	currentInt64  int64
	currentSlice  []any
	currentBool   bool
)

type object struct {
	identifier *string
	fields     []any
}

// Reads bytecode from file and runs it.
func (vm *VirtualMachine) Execute() {
	// Read instructions
	reader := NewInstructionReader(vm.filePath, vm)
	reader.Read()

	// Enter root scope
	vm.stack_scopes[vm.reg_scopeIndex] = filepath.Base(vm.filePath)
	vm.reg_scopeIndex++

	vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	// Interpret instructions
	vm.instructions = &vm.GlobalsInstructions
	for vm.instructionIndex < len(vm.GlobalsInstructions) {
		vm.executeInstruction()
	}

	vm.instructions = &vm.FunctionsInstructions
	vm.instructionIndex = 0
	for vm.instructionIndex < len(vm.FunctionsInstructions) {
		vm.executeInstruction()
	}
}

// Executes current instruction and prints debugging information
func (vm *VirtualMachine) executeInstructionDebug() {
	prevII := vm.instructionIndex

	vm.executeInstruction()

	fmt.Printf("Instruction: %d %s %v\n", prevII, InstructionTypeToString[(*vm.instructions)[prevII].InstructionType], (*vm.instructions)[prevII].InstructionValue)
	fmt.Printf("Stack: %v\n", vm.stack.items[:vm.stack.size])
	fmt.Printf("Return Stack: %v\n", vm.stack_returnIndexes[:vm.reg_returnIndex])
	fmt.Print("ScopeI: " + fmt.Sprintf("%d", vm.reg_scopeIndex) + "; Scope: {")
	fmt.Printf("%s", vm.stack_scopes[0])

	if vm.reg_scopeIndex > 1 {
		for _, scope := range vm.stack_scopes[1:vm.reg_scopeIndex] {
			if len(scope) == 0 {
				fmt.Print(", U")
			} else {
				fmt.Printf(", %s", scope)
			}
		}
	}

	fmt.Println("}")

	if vm.instructionIndex == len(*vm.instructions) {
		fmt.Println("Next: none\n")
		return
	}

	fmt.Printf("Next: %d %s %v\n", vm.instructionIndex, InstructionTypeToString[(*vm.instructions)[vm.instructionIndex].InstructionType], (*vm.instructions)[vm.instructionIndex].InstructionValue)
	fmt.Scanln()
}

// Executes current instruction and advances instructionIndex.
func (vm *VirtualMachine) executeInstruction() {
	instruction := (*vm.instructions)[vm.instructionIndex]

	switch instruction.InstructionType {
	// 1 ARGUMENT INSTRUCTIONS --------------------------------------------------------------------------

	// Halt
	case IT_Halt:
		os.Exit(int(instruction.InstructionValue[0]))

	// Change file
	case IT_FileMarker:
		vm.instructionIndex++

	// Jumps
	case IT_Jump:
		vm.instructionIndex += instruction.InstructionValue[0]

	case IT_JumpIfFalse:
		if !vm.stack.Pop().(bool) {
			vm.instructionIndex += instruction.InstructionValue[0]
		}

	case IT_JumpIfTrue:
		if vm.stack.Pop().(bool) {
			vm.instructionIndex += instruction.InstructionValue[0]
		}

	case IT_JumpBack:
		vm.instructionIndex -= instruction.InstructionValue[0]

	// Call functions
	case IT_Call:
		// Push return address to stack
		vm.stack_returnIndexes[vm.reg_returnIndex] = vm.instructionIndex + 1
		vm.reg_returnIndex++

		// Return address stack overflow
		if vm.reg_returnIndex == STACK_RETURN_INDEX_SIZE {
			vm.panic("Function return address stack overflow.")
		}

		// Jump to function
		vm.instructionIndex = vm.functions[instruction.InstructionValue[0]]
		return

	case IT_CallBuiltInFunc:
		vm.callBuiltInFunction(instruction.InstructionValue[0])

	case IT_PushScope:
		vm.stack_scopes[vm.reg_scopeIndex] = vm.Constants[instruction.InstructionValue[0]].(string)
		vm.reg_scopeIndex++

		if vm.reg_scopeIndex == STACK_SCOPES_SIZE {
			vm.panic("Scope stack overflow. This is probably caused by infinite recursion.")
		}

		vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	// Declare variables
	case IT_DeclareBool:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Bool, nil}, nil}})

	case IT_DeclareInt:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Int, nil}, nil}})

	case IT_DeclareFloat:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Float, nil}, nil}})

	case IT_DeclareString:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_String, nil}, nil}})

	case IT_DeclareList, IT_DeclareSet:
		vm.instructionIndex++

		dataType := data.DataType{declareInstructionToDataType[instruction.InstructionType], nil}
		endType := &dataType.SubType

		for IsCompositeDeclarator((*vm.instructions)[vm.instructionIndex].InstructionType) {
			dataType = data.DataType{declareInstructionToDataType[(*vm.instructions)[vm.instructionIndex].InstructionType], dataType}
			vm.instructionIndex++
		}

		*endType = InstructionToDataType[(*vm.instructions)[vm.instructionIndex].InstructionType]

		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{dataType, []any{}}})

	case IT_DeclareObject:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Object, nil}, nil}})

	case IT_DeclareOption:
		vm.stack_symbolTables.Top.Value.(*SymbolMap).Insert(instruction.InstructionValue[0], &Symbol{ST_Variable, &VariableSymbol{data.DataType{data.DT_Option, nil}, nil}})

	// Set and load list at index
	case IT_SetListAtAToB:
		vm.findSymbol().symbolValue.(*VariableSymbol).value.([]any)[vm.stack.Pop().(int64)] = vm.stack.Pop()

	// Load and store
	case IT_LoadConst:
		vm.stack.Push(vm.Constants[instruction.InstructionValue[0]])

	case IT_LoadConstToList:
		(*vm.stack.Top()) = append((*vm.stack.Top()).([]any), vm.Constants[instruction.InstructionValue[0]])

	case IT_Load:
		vm.stack.Push(vm.findSymbol().symbolValue.(*VariableSymbol).value)

	case IT_Store:
		vm.findSymbol().symbolValue.(*VariableSymbol).value = *vm.stack.Top()

	case IT_StoreAndPop:
		vm.findSymbol().symbolValue.(*VariableSymbol).value = vm.stack.Pop()

	// Objects
	case IT_CreateObject:
		identifier := vm.Constants[instruction.InstructionValue[0]].(string)
		vm.stack.Push(object{&identifier, []any{}})

	case IT_GetField:
		vm.stack.Push(vm.stack.items[vm.stack.size-1].(object).fields[instruction.InstructionValue[0]])

	case IT_GetFieldAndPop:
		vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(object).fields[instruction.InstructionValue[0]]

	case IT_SetField:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1].(object).fields[instruction.InstructionValue[0]] = vm.stack.items[vm.stack.size]

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

	// Logical operations
	case IT_And:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(bool) && vm.stack.items[vm.stack.size].(bool)

	case IT_Or:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(bool) || vm.stack.items[vm.stack.size].(bool)

	// Concatenations
	case IT_StringConcat:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(string) + vm.stack.items[vm.stack.size].(string)

	case IT_ListConcat:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1] = append(vm.stack.items[vm.stack.size-1].([]any), vm.stack.items[vm.stack.size].([]any)...)

	// Return from a function
	case IT_Return:
		vm.stack_symbolTables.Pop()
		vm.reg_scopeIndex--

		if vm.reg_scopeIndex <= 1 {
			os.Exit(0)
		}

		vm.reg_returnIndex--
		vm.instructionIndex = vm.stack_returnIndexes[vm.reg_returnIndex] - 1

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

	// Push literals
	case IT_PushTrue:
		vm.stack.Push(true)

	case IT_PushFalse:
		vm.stack.Push(false)

	case IT_PushNone:
		vm.stack.Push(nil)

	// Scopes
	case IT_PushScopeUnnamed:
		vm.stack_scopes[vm.reg_scopeIndex] = ""
		vm.reg_scopeIndex++

		if vm.reg_scopeIndex == STACK_SCOPES_SIZE {
			vm.panic("Scope stack overflow. This is probably caused by infinite recursion.")
		}

		vm.stack_symbolTables.Push(NewSymbolMap(SYMBOL_MAP_SIZE))

	case IT_PopScope:
		vm.stack_symbolTables.Pop()
		vm.reg_scopeIndex--

	// Adding fields to an object
	case IT_AddField:
		vm.stack.size--

		currentObject, _ = vm.stack.items[vm.stack.size-1].(object)
		currentObject.fields = append(currentObject.fields, vm.stack.items[vm.stack.size])
		vm.stack.items[vm.stack.size-1] = currentObject

	// List operations
	case IT_CreateList:
		vm.stack.Push([]any{})

	case IT_AppendToList:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1] = append(vm.stack.items[vm.stack.size-1].([]any), vm.stack.items[vm.stack.size])

	case IT_IndexList:
		vm.stack.size--

		if int64(len(vm.stack.items[vm.stack.size-1].([]any)))-1 < vm.stack.items[vm.stack.size].(int64) {
			vm.panic(fmt.Sprintf("List index out of range. List size: %d, index: %d.", len(vm.stack.items[vm.stack.size-1].([]any)), vm.stack.items[vm.stack.size].(int64)))
		}

		vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].([]any)[vm.stack.items[vm.stack.size].(int64)]

	case IT_ListContains:
		vm.stack.size--

		currentBool = false
		for _, item := range vm.stack.items[vm.stack.size-1].([]any) {
			if item == vm.stack.items[vm.stack.size] {
				currentBool = true
				break
			}
		}
		vm.stack.items[vm.stack.size-1] = currentBool

	case IT_RemoveListElement:
		vm.stack.size--
		currentInt64 = vm.stack.items[vm.stack.size].(int64)
		currentSlice = vm.stack.items[vm.stack.size-1].([]any)

		if currentInt64 >= int64(len(currentSlice)) {
			vm.panic("List index out of range: index: " + fmt.Sprintf("%d", currentInt64) + ", list size: " + fmt.Sprintf("%d.", len(currentSlice)))
		}

		vm.stack.items[vm.stack.size-1] = append(currentSlice[:currentInt64], currentSlice[currentInt64+1:]...)

	// String operations
	case IT_IndexString:
		vm.stack.size--

		if int64(len(vm.stack.items[vm.stack.size-1].(string)))-1 < vm.stack.items[vm.stack.size].(int64) {
			vm.panic(fmt.Sprintf("String index out of range. Length is %d, index is %d.", len(vm.stack.items[vm.stack.size-1].(string)), vm.stack.items[vm.stack.size].(int64)))
		}

		vm.stack.items[vm.stack.size-1] = string([]rune(vm.stack.items[vm.stack.size-1].(string))[vm.stack.items[vm.stack.size].(int64)])

	// Set operations
	case IT_CreateSet:
		vm.stack.Push(map[any]struct{}{})

	case IT_InsertToSet:
		vm.stack.size--
		vm.stack.items[vm.stack.size-1].(map[any]struct{})[vm.stack.items[vm.stack.size]] = struct{}{}

	case IT_SetContains:
		vm.stack.size--
		_, vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size-1].(map[any]struct{})[vm.stack.items[vm.stack.size]]

	case IT_RemoveSetElement:
		vm.stack.size--
		delete(vm.stack.items[vm.stack.size-1].(map[any]struct{}), vm.stack.items[vm.stack.size])

	// Panic
	case IT_PanicIfNone:
		if vm.stack.items[vm.stack.size-1] == nil {
			vm.panic("Unwrapped option doesn't have a value.")
		}

	// Stack
	case IT_Pop:
		vm.stack.size--

	case IT_DuplicateTop:
		vm.stack.Push(vm.stack.items[vm.stack.size-1])

	// Unpack or default
	case IT_UnpackOrDefault:
		vm.stack.size--
		if vm.stack.items[vm.stack.size-1] == nil {
			vm.stack.items[vm.stack.size-1] = vm.stack.items[vm.stack.size]
		}

	// Ignore line offsets
	case IT_LineOffset:

	// Unknown instruction
	default:
		vm.panic(fmt.Sprintf("Unknown instruction type: %v.", (*vm.instructions)[vm.instructionIndex].InstructionValue))
	}
	vm.instructionIndex++
}

// Finds symbol with identifier from first argument of current instruction. Looks from deepest to shallowest scope symbol table.
// Panics if symbol can't be found.
func (vm *VirtualMachine) findSymbol() *Symbol {
	// Find variable
	symbolTable := vm.stack_symbolTables.Top
	value := symbolTable.Value.(*SymbolMap).Get((*vm.instructions)[vm.instructionIndex].InstructionValue[0])

	for value == nil && symbolTable.Previous != nil {
		symbolTable = symbolTable.Previous
		value = symbolTable.Value.(*SymbolMap).Get((*vm.instructions)[vm.instructionIndex].InstructionValue[0])
	}

	// Couldn't find variable
	if value == nil {
		vm.panic(fmt.Sprintf("Undeclared variable with ID: %v.", (*vm.instructions)[vm.instructionIndex].InstructionValue))
	}

	return value
}

// Prints panic message, traceback and exits program with exit code 1.
func (vm *VirtualMachine) panic(message string) {
	fmt.Println("\033[91mPanic in function " + vm.stack_scopes[vm.reg_scopeIndex-1] + ": " + message + "\033[0m")

	// Get file and line
	lines := map[int]int{}
	file := 0

	for i := 0; i < vm.instructionIndex; i++ {
		if (*vm.instructions)[i].InstructionType == IT_FileMarker {
			file = (*vm.instructions)[i].InstructionValue[0]
			lines[file] = (*vm.instructions)[i+1].InstructionValue[0]
			i++
		} else if (*vm.instructions)[i].InstructionType == IT_LineOffset {
			lines[file] += (*vm.instructions)[i].InstructionValue[0]
		}
	}

	// Put current position on return stack
	vm.stack_returnIndexes[vm.reg_returnIndex] = vm.instructionIndex
	vm.reg_returnIndex++

	vm.traceback()
	os.Exit(1)
}

// Prints all functions on scope stack.
func (vm *VirtualMachine) traceback() {
	fmt.Println("Traceback:")
	for i := vm.reg_returnIndex - 1; i > 0; i-- {
		line, file := vm.traceLine(vm.stack_returnIndexes[i])
		fmt.Printf("   %d file %s.neco, line %d, function %s()\n", i, file, line, vm.stack_scopes[i])
	}
}

// Calculates file and line of instruction at instructionIndex.
func (vm *VirtualMachine) traceLine(instructionIndex int) (int, string) {
	lines := map[int]int{}
	file := 0

	for i := 0; i < instructionIndex; i++ {
		if (*vm.instructions)[i].InstructionType == IT_FileMarker {
			file = (*vm.instructions)[i].InstructionValue[0]
			i++

			lines[file] = (*vm.instructions)[i].InstructionValue[0]
		} else if (*vm.instructions)[i].InstructionType == IT_LineOffset {
			lines[file] += (*vm.instructions)[i].InstructionValue[0]
		}
	}

	return lines[file], vm.Constants[file].(string)
}

var declareInstructionToDataType = map[byte]data.PrimitiveType{
	IT_DeclareList: data.DT_List,
	IT_DeclareSet:  data.DT_Set,
}
