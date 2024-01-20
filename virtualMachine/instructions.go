package virtualMachine

const (
	// 2 arguments
	IT_LoadConstant byte = iota
	IT_Push

	// 1 argument
	IT_CallBuiltInFunction
	IT_Halt
	IT_StoreRegisterA
	IT_LoadRegisterA

	// 0 arguments
	IT_SwapGeneric

	IT_IntAdd
	IT_IntSubtract
	IT_IntMultiply
	IT_IntDivide
	IT_IntPower
	IT_IntModulo

	IT_DeclareBool
	IT_DeclareInt
	IT_DeclareFloat
	IT_DeclareString

	IT_LineOffset
)

var InstructionTypeToString = map[byte]string{
	IT_LoadConstant: "LOAD_CONSTANT",
	IT_Push:         "PUSH",

	IT_CallBuiltInFunction: "CALL_BUILTIN_FUNCTION",
	IT_Halt:                "HALT",
	IT_StoreRegisterA:      "STORE_REGISTER_A",
	IT_LoadRegisterA:       "LOAD_REGISTER_A",

	IT_SwapGeneric: "SWAP_GENERIC",
	IT_IntAdd:      "INT_ADD",
	IT_IntSubtract: "INT_SUBTRACT",
	IT_IntMultiply: "INT_MULTIPLY",
	IT_IntDivide:   "INT_DIVIDE",
	IT_IntPower:    "INT_POWER",
	IT_IntModulo:   "INT_MODULO",

	IT_DeclareBool:   "DECLARE_BOOL",
	IT_DeclareInt:    "DECLARE_INT",
	IT_DeclareFloat:  "DECLARE_FLOAT",
	IT_DeclareString: "DECLARE_STRING",

	IT_LineOffset: "LINE_OFFSET",
}

type Instruction struct {
	InstructionType  byte
	InstructionValue []byte
}
