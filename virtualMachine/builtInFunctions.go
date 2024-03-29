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
	BIF_EnumToInt

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
		vm.stack.Push(string(char))

	// Lengths
	case BIF_StringLength:
		vm.stack.Push(int64(len(vm.stack.Pop().(string))))

	case BIF_ListLength:
		vm.stack.Push(int64(len(vm.stack.Pop().([]any))))

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

func necoPrint(value any, root bool) {
	if _, ok := value.([]any); ok {
		// Print list
		print("{")
		for _, element := range value.([]any)[:len(value.([]any))-1] {
			necoPrint(element, false)
			print(", ")
		}
		necoPrint(value.([]any)[len(value.([]any))-1], false)
		print("}")

	} else if _, ok := value.(string); ok && !root {
		// Print string
		fmt.Printf("\"%v\"", value)
	} else {
		// Use default formatting for everything else
		fmt.Printf("%v", value)
	}
}

func necoPrintString(value any, root bool) string {
	if object, ok := value.(object); ok {
		// Print object
		if len(object.fields) == 0 {
			return "{}"
		}

		str := *object.identifier + "{"

		for _, property := range object.fields[:len(object.fields)-1] {
			str += necoPrintString(property, false) + ", "
		}

		return str + necoPrintString(object.fields[len(object.fields)-1], false) + "}"

	} else if valueList, ok := value.([]any); ok {
		// Print list
		if len(valueList) == 0 {
			return "[]"
		}

		str := "["

		for _, element := range valueList[:len(valueList)-1] {
			str += necoPrintString(element, false) + ", "
		}

		return str + necoPrintString(valueList[len(valueList)-1], false) + "]"

	} else if valueSet, ok := value.(map[any]struct{}); ok {
		// Print set

		str := "{"
		first := true

		for item := range valueSet {
			if first {
				str += necoPrintString(item, false)
				first = false
			} else {
				str += ", " + necoPrintString(item, false)
			}
		}

		return str + "}"

	} else if valueString, ok := value.(string); ok && !root {
		// Print string
		return "\"" + valueString + "\""

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

var BuiltInFuncToString = map[byte]string{
	BIF_Print:     "print",
	BIF_PrintLine: "printLine",

	BIF_AnyToString: "str",

	BIF_BoolToInt: "int",
	BIF_EnumToInt: "int",

	BIF_IntToFloat: "flt",

	BIF_Floor:      "floor",
	BIF_FloorToInt: "floorToInt",
	BIF_Ceil:       "ceil",
	BIF_CeilToInt:  "ceilToInt",
	BIF_Round:      "round",
	BIF_RoundToInt: "roundToInt",
	BIF_AbsInt:     "absInt",
	BIF_AbsFloat:   "absFloat",

	BIF_ReadLine: "readLine",
	BIF_ReadChar: "readChar",

	BIF_StringLength: "length",
	BIF_ListLength:   "size",

	BIF_ToLower: "toLower",
	BIF_ToUpper: "toUpper",

	BIF_RandomInt:      "randomInt",
	BIF_RandomFloat:    "randomFloat",
	BIF_RandomRangeInt: "randomRangeInt",

	BIF_Trace: "trace",
}
