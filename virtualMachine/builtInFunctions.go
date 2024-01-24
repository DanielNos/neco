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

	BIF_ReadLine
	BIF_ReadChar
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
			fmt.Printf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(string))
		}

	case BIF_PrintLine:
		if vm.Stack_Argument[vm.Reg_ArgumentPointer-1] == nil {
			println("none")
		} else {
			fmt.Printf("%v\n", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(string))
		}
		vm.Reg_ArgumentPointer--

	// Data types to string
	case BIF_BoolToString:
		vm.Reg_GenericA = fmt.Sprintf("%v", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(bool))
		vm.Reg_ArgumentPointer--

	case BIF_IntToString:
		vm.Reg_GenericA = fmt.Sprintf("%d", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(int64))
		vm.Reg_ArgumentPointer--

	case BIF_FloatToString:
		vm.Reg_GenericA = fmt.Sprintf("%f", vm.Stack_Argument[vm.Reg_ArgumentPointer-1].(float64))
		vm.Reg_ArgumentPointer--

	// Data type to data type
	case BIF_BoolToInt:
		if vm.Reg_GenericA.(bool) {
			vm.Reg_GenericA = INT_1
		} else {
			vm.Reg_GenericA = INT_0
		}
		vm.Reg_ArgumentPointer--

	case BIF_IntToFloat:
		vm.Reg_GenericA = float64(vm.Reg_GenericA.(int64))
		vm.Reg_ArgumentPointer--

	// Rounding floats
	case BIF_Floor:
		vm.Reg_GenericA = math.Floor(vm.Reg_GenericA.(float64))
		vm.Reg_ArgumentPointer--

	case BIF_FloorToInt:
		vm.Reg_GenericA = int64(vm.Reg_GenericA.(float64))
		vm.Reg_ArgumentPointer--

	case BIF_Ceil:
		vm.Reg_GenericA = math.Ceil(vm.Reg_GenericA.(float64))
		vm.Reg_ArgumentPointer--

	case BIF_CeilToInt:
		vm.Reg_GenericA = int64(math.Ceil(vm.Reg_GenericA.(float64)))
		vm.Reg_ArgumentPointer--

	case BIF_Round:
		vm.Reg_GenericA = math.Round(vm.Reg_GenericA.(float64))
		vm.Reg_ArgumentPointer--

	case BIF_RoundToInt:
		vm.Reg_GenericA = int64(math.Round(vm.Reg_GenericA.(float64)))
		vm.Reg_ArgumentPointer--

	// Reading from terminal
	case BIF_ReadLine:
		vm.Reg_GenericA, _ = vm.reader.ReadString('\n')
		vm.Reg_GenericA = vm.Reg_GenericA.(string)[:len(vm.Reg_GenericA.(string))-1]

	case BIF_ReadChar:
		char, _, _ := vm.reader.ReadRune()
		vm.Reg_GenericA = fmt.Sprintf("%c", char)
	}
}
