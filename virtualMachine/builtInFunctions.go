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

	BIF_AnyToString

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

	BIF_StringLength
	BIF_ListLength

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
		necoPrint(vm.stack.Pop(), true)

	case BIF_PrintLine:
		necoPrint(vm.stack.Pop(), true)
		println()

	// Data types to string
	case BIF_AnyToString:
		vm.stack.Push(necoPrintString(vm.stack.Pop(), true))

	// Data type to data type
	case BIF_BoolToInt:
		if vm.stack.Pop().(bool) {
			vm.stack.Push(INT_1)
		} else {
			vm.stack.Push(INT_0)
		}

	case BIF_IntToFloat:
		vm.stack.Push(float64(vm.stack.Pop().(int64)))

	// Rounding floats
	case BIF_Floor:
		vm.stack.Push(math.Floor(vm.stack.Pop().(float64)))

	case BIF_FloorToInt:
		vm.stack.Push(int64(vm.stack.Pop().(float64)))

	case BIF_Ceil:
		vm.stack.Push(math.Ceil(vm.stack.Pop().(float64)))

	case BIF_CeilToInt:
		vm.stack.Push(int64(math.Ceil(vm.stack.Pop().(float64))))

	case BIF_Round:
		vm.stack.Push(math.Round(vm.stack.Pop().(float64)))

	case BIF_RoundToInt:
		vm.stack.Push(int64(math.Round(vm.stack.Pop().(float64))))

	// Absolute values
	case BIF_AbsInt:
		if (*vm.stack.Top()).(int64) < 0 {
			vm.stack.Push(-vm.stack.Pop().(int64))
		}

	case BIF_AbsFloat:
		vm.stack.Push(math.Abs(vm.stack.Pop().(float64)))

	// Reading from terminal
	case BIF_ReadLine:
		line, _ := vm.reader.ReadString('\n')
		vm.stack.Push(line[:len(line)-1])

	case BIF_ReadChar:
		char, _, _ := vm.reader.ReadRune()
		vm.stack.Push(fmt.Sprintf("%c", char))

	// Lengths
	case BIF_StringLength:
		vm.stack.Push(int64(len(vm.stack.Pop().(string))))

	case BIF_ListLength:
		vm.stack.Push(int64(len(vm.stack.Pop().([]interface{}))))

	// String functions
	case BIF_ToLower:
		vm.stack.Push(strings.ToLower(vm.stack.Pop().(string)))

	case BIF_ToUpper:
		vm.stack.Push(strings.ToUpper(vm.stack.Pop().(string)))

	// Random numbers
	case BIF_RandomInt:
		vm.stack.Push(int64(rand.Uint64()))

	case BIF_RandomFloat:
		vm.stack.Push(rand.Float64())

	case BIF_RandomRangeInt:
		vm.stack.Push(rand.Int63n(vm.stack.Pop().(int64)-(*vm.stack.Top()).(int64)+1) + vm.stack.Pop().(int64))

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

func necoPrint(value interface{}, root bool) {
	if _, ok := value.([]interface{}); ok {
		// Print list
		print("{")
		for _, element := range value.([]interface{})[:len(value.([]interface{}))-1] {
			necoPrint(element, false)
			print(", ")
		}
		necoPrint(value.([]interface{})[len(value.([]interface{}))-1], false)
		print("}")

	} else if _, ok := value.(string); ok && !root {
		// Print string
		fmt.Printf("\"%v\"", value)
	} else {
		// Use default formatting for everything else
		fmt.Printf("%v", value)
	}
}

func necoPrintString(value interface{}, root bool) string {
	if _, ok := value.([]interface{}); ok {
		// Print list
		str := "{"
		for _, element := range value.([]interface{})[:len(value.([]interface{}))-1] {
			str = fmt.Sprintf("%s%s, ", str, necoPrintString(element, false))
		}
		str = fmt.Sprintf("%s%s}", str, necoPrintString(value.([]interface{})[len(value.([]interface{}))-1], false))

		return str

	} else if _, ok := value.(string); ok && !root {
		// Print string
		return fmt.Sprintf("\"%v\"", value)
	} else {
		// Use default formatting for everything else
		return fmt.Sprintf("%v", value)
	}
}

func PowerInt64(base, exponent int64) int64 {
	var result int64 = 1

	for exponent > 0 {
		if exponent%2 == 1 {
			result *= base
		}
		base *= base
		exponent /= 2
	}

	return result
}
