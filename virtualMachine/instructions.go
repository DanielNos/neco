package virtualMachine

const (
	// 1 argument
	IT_CallBuiltInFunc byte = iota
	IT_Halt

	IT_StoreRegA
	IT_StoreRegB

	IT_LoadConstRegA
	IT_LoadConstRegB

	IT_LoadRegA
	IT_LoadRegB

	// 0 arguments
	IT_SwapGeneric

	IT_PushRegAToArgStack
	IT_PushRegBToArgStack

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
	IT_CallBuiltInFunc: "CALL_BUILTIN_FUNC",
	IT_Halt:            "HALT",

	IT_StoreRegA: "STORE_REG_A",
	IT_StoreRegB: "STORE_REG_B",

	IT_LoadConstRegA: "LOAD_CONST_REG_A",
	IT_LoadConstRegB: "LOAD_CONST_REG_B",

	IT_LoadRegA: "LOAD_REG_A",
	IT_LoadRegB: "LOAD_REG_B",

	// 0 arguments
	IT_SwapGeneric: "SWAP_GENERIC",

	IT_PushRegAToArgStack: "PUSH_REG_A_ARG_STACK",
	IT_PushRegBToArgStack: "PUSH_REG_B_ARG_STACK",

	IT_IntAdd:      "INT_ADD",
	IT_IntSubtract: "INT_SUB",
	IT_IntMultiply: "INT_MUL",
	IT_IntDivide:   "INT_DIV",
	IT_IntPower:    "INT_POW",
	IT_IntModulo:   "INT_MOD",

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
