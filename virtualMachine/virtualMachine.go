package virtualMachine

import "fmt"

const (
	VERSION_MAJOR byte = 0
	VERSION_MINOR      = 1
	VERSION_PATCH      = 0
)

const (
	Reg_IntegerA byte = iota
	Reg_IntegerB
	Reg_FloatA
	Reg_FloatB
	Stack_Argument
)

const STACK_ARGUMENT_SIZE = 100

type VirtualMachine struct {
	Constants    []interface{}
	Instructions []Instruction

	Reg_IntegerA int64
	Reg_IntegerB int64

	Reg_FloatA float64
	Reg_FloatB float64

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
			case Reg_IntegerA:
				vm.Reg_IntegerA = vm.Constants[instruction.InstructionValue[0]].(int64)
			case Reg_IntegerB:
				vm.Reg_IntegerB = vm.Constants[instruction.InstructionValue[0]].(int64)

			case Reg_FloatA:
				vm.Reg_FloatA = vm.Constants[instruction.InstructionValue[0]].(float64)
			case Reg_FloatB:
				vm.Reg_FloatB = vm.Constants[instruction.InstructionValue[0]].(float64)

			case Stack_Argument:
				vm.Stack_Argument[vm.Reg_ArgumentPointer] = vm.Constants[instruction.InstructionValue[0]]
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
	}
}
