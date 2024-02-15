package virtualMachine

const (
	// 1 argument 2 bytes
	IT_JumpEx byte = iota
	IT_JumpIfTrueEx

	// 1 argument 1 byte
	IT_CallBuiltInFunc
	IT_Halt

	IT_DeclareBool
	IT_DeclareInt
	IT_DeclareFloat
	IT_DeclareString
	IT_DeclareList

	IT_StoreRegA
	IT_StoreRegB

	IT_LoadConstRegA
	IT_LoadConstRegB

	IT_LoadRegA
	IT_LoadRegB

	IT_PushScope

	IT_PopArgStackVariable

	IT_Call
	IT_JumpBack
	IT_Jump
	IT_JumpIfTrue

	// 0 arguments
	IT_SwapAB
	IT_CopyRegAToC
	IT_CopyRegBToC
	IT_CopyRegCToA
	IT_CopyRegCToB
	IT_CopyRegAToD
	IT_CopyRegDToA
	IT_CopyRegDToB

	IT_PushRegAToArgStack
	IT_PushRegBToArgStack
	IT_PopArgStackRegA

	IT_IntAdd
	IT_IntSubtract
	IT_IntMultiply
	IT_IntDivide
	IT_IntPower
	IT_IntModulo

	IT_FloatAdd
	IT_FloatSubtract
	IT_FloatMultiply
	IT_FloatDivide
	IT_FloatPower
	IT_FloatModulo

	IT_StringConcat

	IT_Return

	IT_Equal
	IT_LowerInt
	IT_LowerFloat
	IT_GreaterInt
	IT_GreaterFloat
	IT_LowerEqualInt
	IT_LowerEqualFloat
	IT_GreaterEqualInt
	IT_GreaterEqualFloat
	IT_Not

	IT_SetRegAFalse
	IT_SetRegATrue
	IT_SetRegBFalse
	IT_SetRegBTrue

	IT_PushScopeUnnamed
	IT_PopScope

	IT_LineOffset
)

var InstructionTypeToString = map[byte]string{
	// 1 argument 2 bytes
	IT_JumpEx:       "JUMP_EX",
	IT_JumpIfTrueEx: "JUMP_IF_TRUE_EX",

	// 1 argument 1 byte
	IT_CallBuiltInFunc: "CALL_BUILTIN_FUNC",
	IT_Halt:            "HALT",

	IT_StoreRegA: "STORE_REG_A",
	IT_StoreRegB: "STORE_REG_B",

	IT_LoadConstRegA: "LOAD_CONST_REG_A",
	IT_LoadConstRegB: "LOAD_CONST_REG_B",

	IT_LoadRegA: "LOAD_REG_A",
	IT_LoadRegB: "LOAD_REG_B",

	IT_PushScope: "PUSH_SCOPE",

	IT_PopArgStackVariable: "POP_ARG_STACK_VARIABLE",

	IT_Call:       "CALL",
	IT_JumpBack:   "JUMP_BACK",
	IT_Jump:       "JUMP",
	IT_JumpIfTrue: "JUMP_IF_TRUE",

	// 0 arguments
	IT_SwapAB:      "SWAP_A_B",
	IT_CopyRegAToC: "COPY_REG_A_TO_C",
	IT_CopyRegBToC: "COPY_REG_B_TO_C",
	IT_CopyRegCToA: "COPY_REG_C_TO_A",
	IT_CopyRegCToB: "COPY_REG_C_TO_B",
	IT_CopyRegAToD: "COPY_REG_A_TO_D",
	IT_CopyRegDToA: "COPY_REG_D_TO_A",
	IT_CopyRegDToB: "COPY_REG_D_TO_B",

	IT_PushRegAToArgStack: "PUSH_REG_A_ARG_STACK",
	IT_PushRegBToArgStack: "PUSH_REG_B_ARG_STACK",
	IT_PopArgStackRegA:    "POP_ARG_STACK_REG_A",

	IT_IntAdd:      "INT_ADD",
	IT_IntSubtract: "INT_SUB",
	IT_IntMultiply: "INT_MUL",
	IT_IntDivide:   "INT_DIV",
	IT_IntPower:    "INT_POW",
	IT_IntModulo:   "INT_MOD",

	IT_FloatAdd:      "FLT_ADD",
	IT_FloatSubtract: "FLT_SUB",
	IT_FloatMultiply: "FLT_MUL",
	IT_FloatDivide:   "FLT_DIV",
	IT_FloatPower:    "FLT_POW",
	IT_FloatModulo:   "FLT_MOD",

	IT_StringConcat: "STRING_CONCAT",

	IT_DeclareBool:   "DECLARE_BOOL",
	IT_DeclareInt:    "DECLARE_INT",
	IT_DeclareFloat:  "DECLARE_FLOAT",
	IT_DeclareString: "DECLARE_STRING",
	IT_DeclareList:   "DECLARE_LIST",

	IT_Return: "RETURN",

	IT_Equal:             "EQUAL",
	IT_LowerInt:          "LOWER_INT",
	IT_GreaterInt:        "GREATER_INT",
	IT_LowerEqualInt:     "LOWER_EQUAL_INT",
	IT_GreaterEqualInt:   "GREATER_EQUAL_INT",
	IT_LowerFloat:        "LOWER_FLOAT",
	IT_GreaterFloat:      "GREATER_FLOAT",
	IT_LowerEqualFloat:   "LOWER_EQUAL_FLOAT",
	IT_GreaterEqualFloat: "GREATER_EQUAL_FLOAT",
	IT_Not:               "NOT",

	IT_SetRegAFalse: "SET_REG_A_FALSE",
	IT_SetRegATrue:  "SET_REG_A_TRUE",
	IT_SetRegBFalse: "SET_REG_B_FALSE",
	IT_SetRegBTrue:  "SET_REG_B_TRUE",

	IT_PushScopeUnnamed: "PUSH_SCOPE_UNNAMED",
	IT_PopScope:         "POP_SCOPE",

	IT_LineOffset: "LINE_OFFSET",
}

type Instruction struct {
	InstructionType  byte
	InstructionValue []byte
}

type ExpandedInstruction struct {
	InstructionType  byte
	InstructionValue []int
}
