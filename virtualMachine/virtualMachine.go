package virtualMachine

import "fmt"

const (
	Reg_Operation0A byte = iota
	Reg_Operation0B
)

const (
	Stack_Argument byte = iota
)

const STACK_ARGUMENT_SIZE = 100

type VirtualMachine struct {
	Constants    []interface{}
	Instructions []Instruction

	Reg_Operation0A int64
	Reg_Operation0B int64

	Reg_ArgumentPointer int
	Stack_Argument      []interface{}
}

func NewVirutalMachine() *VirtualMachine {
	return &VirtualMachine{Stack_Argument: make([]interface{}, STACK_ARGUMENT_SIZE)}
}

func (vm *VirtualMachine) Execute() {
	for _, instruction := range vm.Instructions {
		switch instruction.InstructionType {
		case IT_LoadConstant:
			switch instruction.ValueB {
			case Stack_Argument:
				vm.Stack_Argument[vm.Reg_ArgumentPointer] = vm.Constants[instruction.ValueA.(byte)]
				vm.Reg_ArgumentPointer++
			}
		case IT_CallBuiltInFunction:
			vm.callBuiltInFunction(instruction.ValueA.(byte))
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
