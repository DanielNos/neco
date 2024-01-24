package virtualMachine

import (
	"fmt"
	"math"
)

const (
	BIF_Print byte = iota
	BIF_PrintLine

	BIF_BoolToString
	BIF_IntToString
	BIF_FloatToString

	BIF_BoolToInt
	BIF_IntToFloat

	BIF_Floor
	BIF_FloorToInt
	BIF_Ceil
	BIF_CeilToInt
	BIF_Round
	BIF_RoundToInt
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

	case BIF_PrintLine:
		if vm.Stack_Argument[vm.Reg_ArgumentPointer-1] == nil {
			println("none")
		} else {
			fmt.Printf("%v\n", vm.Stack_Argument[vm.Reg_ArgumentPointer-1])
		}

	// Data types to string
	case BIF_BoolToString:
		vm.Reg_GenericA = fmt.Sprintf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(bool))

	case BIF_IntToString:
		vm.Reg_GenericA = fmt.Sprintf("%d", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(int64))

	case BIF_FloatToString:
		vm.Reg_GenericA = fmt.Sprintf("%f", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(float64))

	// Data type to data type
	case BIF_BoolToInt:
		if vm.Reg_GenericA.(bool) {
			vm.Reg_GenericA = INT_1
		} else {
			vm.Reg_GenericA = INT_0
		}

	case BIF_IntToFloat:
		vm.Reg_GenericA = float64(vm.Reg_GenericA.(int64))

	// Rounding floats
	case BIF_Floor:
		vm.Reg_GenericA = math.Floor(vm.Reg_GenericA.(float64))

	case BIF_FloorToInt:
		vm.Reg_GenericA = int64(vm.Reg_GenericA.(float64))

	case BIF_Ceil:
		vm.Reg_GenericA = math.Ceil(vm.Reg_GenericA.(float64))

	case BIF_CeilToInt:
		vm.Reg_GenericA = int64(math.Ceil(vm.Reg_GenericA.(float64)))

	case BIF_Round:
		vm.Reg_GenericA = math.Round(vm.Reg_GenericA.(float64))

	case BIF_RoundToInt:
		vm.Reg_GenericA = int64(math.Round(vm.Reg_GenericA.(float64)))
	}

	vm.Reg_ArgumentPointer--
}
