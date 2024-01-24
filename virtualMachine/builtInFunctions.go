package virtualMachine

import "fmt"

const (
	BIF_Print byte = iota
	BIF_PrintLine
	BIF_Bool2String
	BIF_Int2String
	BIF_Float2String
	BIF_Bool2Int
	BIF_Int2Float
)

const INT_0 = int64(0)
const INT_1 = int64(1)

func (vm *VirtualMachine) callBuiltInFunction(functionCode byte) {
	switch functionCode {
	// Print functions
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

	// Data types to string
	case BIF_Bool2String:
		vm.Reg_GenericA = fmt.Sprintf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(bool))
		vm.Reg_ArgumentPointer--

	case BIF_Int2String:
		vm.Reg_GenericA = fmt.Sprintf("%d", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(int64))
		vm.Reg_ArgumentPointer--

	case BIF_Float2String:
		vm.Reg_GenericA = fmt.Sprintf("%f", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(float64))
		vm.Reg_ArgumentPointer--

	// Data type to data type
	case BIF_Bool2Int:
		if vm.Reg_GenericA.(bool) {
			vm.Reg_GenericA = INT_1
		} else {
			vm.Reg_GenericA = INT_0
		}

	case BIF_Int2Float:
		vm.Reg_GenericA = float64(vm.Reg_GenericA.(int64))
	}
}
