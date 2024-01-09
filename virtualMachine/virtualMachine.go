package virtualMachine

import "fmt"

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
}

func NewVirutalMachine() *VirtualMachine {
	return &VirtualMachine{Stack_Argument: make([]interface{}, STACK_ARGUMENT_SIZE)}
}

func (vm *VirtualMachine) Execute(filePath string) {
	reader := NewInstructionReader(filePath, &vm.Instructions, &vm.Constants)
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

		// Add operator
		case IT_IntAdd:

		// Cal Built-In Function
		case IT_CallBuiltInFunction:
			vm.callBuiltInFunction(instruction.InstructionValue[0])
		}
	}
}

func (vm *VirtualMachine) callBuiltInFunction(functionCode byte) {
	switch functionCode {
	case BIF_Print:
		fmt.Printf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1])
		vm.Reg_ArgumentPointer--
	case BIF_PrintLine:
		fmt.Printf("%v\n", vm.Stack_Argument[vm.Reg_ArgumentPointer-1])
		vm.Reg_ArgumentPointer--
	case BIF_Bool2String:
		vm.Reg_GenericA = fmt.Sprintf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(bool))
		vm.Reg_ArgumentPointer--
	case BIF_Int2String:
		vm.Reg_GenericA = fmt.Sprintf("%d", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(int64))
		vm.Reg_ArgumentPointer--
	case BIF_Float2String:
		vm.Reg_GenericA = fmt.Sprintf("%f", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(float64))
		vm.Reg_ArgumentPointer--
	}
}
