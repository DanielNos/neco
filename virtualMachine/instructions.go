package virtualMachine

const (
	// 1 argument 2 bytes
	IT_JumpEx byte = iota
	IT_JumpIfTrueEx

	// 1 argument 1 byte
	IT_LoadListValueRegA
	IT_LoadListValueRegB
	IT_CallBuiltInFunc
	IT_Halt

	IT_DeclareBool
	IT_DeclareInt
	IT_DeclareFloat
	IT_DeclareString
	IT_DeclareList

	IT_StoreRegA
	IT_StoreRegB

	IT_AppendListRegA
	IT_SetListAtAToB

	IT_LoadConstRegA
	IT_LoadConstRegB
	IT_LoadConstArgStack

	IT_LoadRegA
	IT_LoadRegB
	IT_LoadArgStack

	IT_PushScope

	IT_PopArgStackToVariable

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
	IT_CopyRegAToE
	IT_CopyRegEToA
	IT_CopyRegEToB

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
	IT_IntLower
	IT_FloatLower
	IT_IntGreater
	IT_FloatGreater
	IT_IntLowerEqual
	IT_FloatLowerEqual
	IT_IntGreaterEqual
	IT_FloatGreaterEqual
	IT_Not

	IT_SetRegAFalse
	IT_SetRegATrue
	IT_SetRegBFalse
	IT_SetRegBTrue

	IT_PushScopeUnnamed
	IT_PopScope

	IT_CreateListRegE
	IT_AppendRegAListE

	IT_LineOffset
)

var InstructionTypeToString = map[byte]string{
	// 2 arguments 2 bytes
	IT_LoadListValueRegA: "LOAD_LIST_VAL_REG_A",
	IT_LoadListValueRegB: "LOAD_LIST_VAL_REG_A",

	// 1 argument 2 bytes
	IT_JumpEx:       "JUMP_EX",
	IT_JumpIfTrueEx: "JUMP_IF_TRUE_EX",

	// 1 argument 1 byte
	IT_CallBuiltInFunc: "CALL_BUILTIN",
	IT_Halt:            "HALT",

	IT_StoreRegA: "STORE_REG_A",
	IT_StoreRegB: "STORE_REG_B",

	IT_AppendListRegA: "APPEND_LIST_REG_A",
	IT_SetListAtAToB:  "SET_LIST_AT_A_TO_B",

	IT_LoadConstRegA:     "LOAD_CONST_A",
	IT_LoadConstRegB:     "LOAD_CONST_B",
	IT_LoadConstArgStack: "LOAD_CONST_ARG",

	IT_LoadRegA:     "LOAD_A",
	IT_LoadRegB:     "LOAD_B",
	IT_LoadArgStack: "LOAD_ARG",

	IT_PushScope: "PUSH_SCOPE",

	IT_PopArgStackToVariable: "POP_ARG_TO_VAR",

	IT_Call:       "CALL",
	IT_JumpBack:   "JUMP_BACK",
	IT_Jump:       "JUMP",
	IT_JumpIfTrue: "JUMP_IF_TRUE",

	// 0 arguments
	IT_SwapAB:      "SWAP_A_B",
	IT_CopyRegAToC: "COPY_A_TO_C",
	IT_CopyRegBToC: "COPY_B_TO_C",
	IT_CopyRegCToA: "COPY_C_TO_A",
	IT_CopyRegCToB: "COPY_C_TO_B",
	IT_CopyRegAToD: "COPY_A_TO_D",
	IT_CopyRegDToA: "COPY_D_TO_A",
	IT_CopyRegDToB: "COPY_D_TO_B",
	IT_CopyRegAToE: "COPY_A_TO_E",
	IT_CopyRegEToA: "COPY_E_TO_A",
	IT_CopyRegEToB: "COPY_E_TO_B",

	IT_PushRegAToArgStack: "PUSH_A_TO_ARG",
	IT_PushRegBToArgStack: "PUSH_B_TO_ARG",
	IT_PopArgStackRegA:    "POP_ARG_TO_A",

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

	IT_DeclareBool:   "DECL_BOOL",
	IT_DeclareInt:    "DECL_INT",
	IT_DeclareFloat:  "DECL_FLOAT",
	IT_DeclareString: "DECL_STRING",
	IT_DeclareList:   "DECL_LIST",

	IT_Return: "RETURN",

	IT_Equal:             "EQUAL",
	IT_IntLower:          "INT_LOWER",
	IT_IntGreater:        "INT_GREATER",
	IT_IntLowerEqual:     "INT_LOWER_EQUAL",
	IT_IntGreaterEqual:   "INT_GREATER_EQUAL",
	IT_FloatLower:        "FLOAT_LOWER",
	IT_FloatGreater:      "FLOAT_GREATER",
	IT_FloatLowerEqual:   "FLOAT_LOWER_EQUAL",
	IT_FloatGreaterEqual: "FLOAT_GREATER_EQUAL",
	IT_Not:               "NOT",

	IT_SetRegAFalse: "SET_A_FALSE",
	IT_SetRegATrue:  "SET_A_TRUE",
	IT_SetRegBFalse: "SET_B_FALSE",
	IT_SetRegBTrue:  "SET_B_TRUE",

	IT_PushScopeUnnamed: "PUSH_SCOPE_UNNAMED",
	IT_PopScope:         "POP_SCOPE",

	IT_CreateListRegE:  "NEW_LIST_E",
	IT_AppendRegAListE: "APPEND_A_TO_E",

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
