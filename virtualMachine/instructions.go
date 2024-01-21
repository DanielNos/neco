package virtualMachine

const (
	// 1 argument
	IT_CallBuiltInFunction byte = iota
	IT_Halt

	IT_StoreRegisterA
	IT_StoreRegisterB

	IT_LoadConstantRegisterA
	IT_LoadConstantRegisterB

	IT_LoadRegisterA
	IT_LoadRegisterB

	// 0 arguments
	IT_SwapGeneric

	IT_PushRegisterAArgStack
	IT_PushRegisterBArgStack

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
	// 1 argument
	IT_CallBuiltInFunction: "CALL_BUILTIN_FUNCTION",
	IT_Halt:                "HALT",

	IT_StoreRegisterA: "STORE_REGISTER_A",
	IT_StoreRegisterB: "STORE_REGISTER_B",

	IT_LoadConstantRegisterA: "LOAD_CONSTANT_REGISTER_A",
	IT_LoadConstantRegisterB: "LOAD_CONSTANT_REGISTER_B",

	IT_LoadRegisterA: "LOAD_REGISTER_A",
	IT_LoadRegisterB: "LOAD_REGISTER_B",

	// 0 arguments
	IT_SwapGeneric: "SWAP_GENERIC",

	IT_PushRegisterAArgStack: "PUSH_REGISTER_A_ARG_STACK",
	IT_PushRegisterBArgStack: "PUSH_REGISTER_B_ARG_STACK",

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
