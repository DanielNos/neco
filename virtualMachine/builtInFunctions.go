package virtualMachine

import "fmt"

const (
	BIF_Print byte = iota
	BIF_PrintLine
	BIF_Bool2String
	BIF_Int2String
	BIF_Float2String
)

func (vm *VirtualMachine) callBuiltInFunction(functionCode byte) {
	switch functionCode {
	case BIF_Print:
		if vm.Stack_Argument[vm.Reg_ArgumentPointer-1] == nil {
			print("none")
		} else {
			fmt.Printf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1])
		}
		vm.Reg_ArgumentPointer--

	case BIF_PrintLine:
		if vm.Stack_Argument[vm.Reg_ArgumentPointer-1] == nil {
			println("none")
		} else {
			fmt.Printf("%v\n", vm.Stack_Argument[vm.Reg_ArgumentPointer-1])
		}
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
