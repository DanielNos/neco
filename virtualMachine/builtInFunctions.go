package virtualMachine

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const (
	BIF_Print = iota
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
	BIF_AbsInt
	BIF_AbsFloat

	BIF_ReadLine
	BIF_ReadChar

	BIF_Length
	BIF_ToLower
	BIF_ToUpper

	BIF_RandomInt
	BIF_RandomFloat
	BIF_RandomRangeInt

	BIF_Trace
)

const INT_0 = int64(0)
const INT_1 = int64(1)

func (vm *VirtualMachine) callBuiltInFunction(functionCode int) {
	switch functionCode {
	// Print functions
	case BIF_Print:
		vm.reg_argumentPointer--
		fmt.Printf("%v", vm.stack_arguments[vm.reg_argumentPointer])

	case BIF_PrintLine:
		vm.reg_argumentPointer--
		fmt.Printf("%v\n", vm.stack_arguments[vm.reg_argumentPointer])

	// Data types to string
	case BIF_BoolToString:
		vm.reg_argumentPointer--
		vm.reg_genericA = fmt.Sprintf("%v", vm.stack_arguments[vm.reg_argumentPointer].(bool))

	case BIF_IntToString:
		vm.reg_argumentPointer--
		vm.reg_genericA = fmt.Sprintf("%d", vm.stack_arguments[vm.reg_argumentPointer].(int64))

	case BIF_FloatToString:
		vm.reg_argumentPointer--
		vm.reg_genericA = fmt.Sprintf("%f", vm.stack_arguments[vm.reg_argumentPointer].(float64))

	// Data type to data type
	case BIF_BoolToInt:
		if vm.reg_genericA.(bool) {
			vm.reg_genericA = INT_1
		} else {
			vm.reg_genericA = INT_0
		}
		vm.reg_argumentPointer--

	case BIF_IntToFloat:
		vm.reg_genericA = float64(vm.reg_genericA.(int64))
		vm.reg_argumentPointer--

	// Rounding floats
	case BIF_Floor:
		vm.reg_genericA = math.Floor(vm.reg_genericA.(float64))
		vm.reg_argumentPointer--

	case BIF_FloorToInt:
		vm.reg_genericA = int64(vm.reg_genericA.(float64))
		vm.reg_argumentPointer--

	case BIF_Ceil:
		vm.reg_genericA = math.Ceil(vm.reg_genericA.(float64))
		vm.reg_argumentPointer--

	case BIF_CeilToInt:
		vm.reg_genericA = int64(math.Ceil(vm.reg_genericA.(float64)))
		vm.reg_argumentPointer--

	case BIF_Round:
		vm.reg_genericA = math.Round(vm.reg_genericA.(float64))
		vm.reg_argumentPointer--

	case BIF_RoundToInt:
		vm.reg_genericA = int64(math.Round(vm.reg_genericA.(float64)))
		vm.reg_argumentPointer--

	// Absolute values
	case BIF_AbsInt:
		if vm.reg_genericA.(int64) < 0 {
			vm.reg_genericA = -vm.reg_genericA.(int64)
		}

	case BIF_AbsFloat:
		vm.reg_genericA = math.Abs(vm.reg_genericA.(float64))

	// Reading from terminal
	case BIF_ReadLine:
		vm.reg_genericA, _ = vm.reader.ReadString('\n')
		vm.reg_genericA = vm.reg_genericA.(string)[:len(vm.reg_genericA.(string))-1]

	case BIF_ReadChar:
		char, _, _ := vm.reader.ReadRune()
		vm.reg_genericA = fmt.Sprintf("%c", char)

	// String functions
	case BIF_Length:
		vm.reg_genericA = len(vm.reg_genericA.(string))
		vm.reg_argumentPointer--

	case BIF_ToLower:
		vm.reg_genericA = strings.ToLower(vm.reg_genericA.(string))
		vm.reg_argumentPointer--

	case BIF_ToUpper:
		vm.reg_genericA = strings.ToUpper(vm.reg_genericA.(string))
		vm.reg_argumentPointer--

	// Random numbers
	case BIF_RandomInt:
		vm.reg_genericA = int64(rand.Uint64())

	case BIF_RandomFloat:
		vm.reg_genericA = rand.Float64()

	case BIF_RandomRangeInt:
		vm.reg_argumentPointer -= 2
		vm.reg_genericA = rand.Int63n(vm.stack_arguments[vm.reg_argumentPointer+1].(int64)-vm.stack_arguments[vm.reg_argumentPointer].(int64)+1) + vm.stack_arguments[vm.reg_argumentPointer].(int64)

	// Trace
	case BIF_Trace:
		print("[")
		for _, scope := range vm.stack_scopes[:vm.reg_scopeIndex-1] {
			fmt.Printf("\"%v\", ", scope)
		}
		fmt.Printf("\"%v\"", vm.stack_scopes[vm.reg_scopeIndex-1])
		println("]")
	}
}
