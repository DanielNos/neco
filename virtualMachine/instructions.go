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
	IT_SwapOperation
	IT_CopyOpAToOpStore
	IT_CopyOpBToOpStore
	IT_CopyOpStoreToOpA
	IT_CopyOpStoreToOpB
	IT_CopyOpAToReturn
	IT_CopyReturnToOpA
	IT_CopyReturnToOpB
	IT_CopyOpAToListA
	IT_CopyListAToOpA
	IT_CopyListAToOpB

	IT_PushOpAToArg
	IT_PushOpBToArg
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

	IT_CreateListInListA
	IT_AppendOpAToListA

	IT_LineOffset
)

var InstructionTypeToString = map[byte]string{
	// 2 arguments 2 bytes
	IT_LoadListValueRegA: "load_list_at_OA_to_OA",
	IT_LoadListValueRegB: "load_list_at_OB_to_OB",

	// 1 argument 2 bytes
	IT_JumpEx:       "jump_ex",
	IT_JumpIfTrueEx: "jump_if_true_ex",

	// 1 argument 1 byte
	IT_CallBuiltInFunc: "call_builtin",
	IT_Halt:            "halt",

	IT_StoreRegA: "store_OA",
	IT_StoreRegB: "store_OB",

	IT_SetListAtAToB: "set_list_at_OA_to_OB",

	IT_LoadConstRegA:     "load_const_OA",
	IT_LoadConstRegB:     "load_const_OB",
	IT_LoadConstArgStack: "load_const_ARG",

	IT_LoadRegA:     "load_OA",
	IT_LoadRegB:     "load_OB",
	IT_LoadArgStack: "load_ARG",

	IT_PushScope: "push_scope",

	IT_PopArgStackToVariable: "pop_ARG_store",

	IT_Call:       "call",
	IT_JumpBack:   "jump_back",
	IT_Jump:       "jump",
	IT_JumpIfTrue: "jump_if_true",

	// 0 arguments
	IT_SwapOperation:    "swap_OA_OB",
	IT_CopyOpAToOpStore: "copy_OA_to_OS",
	IT_CopyOpBToOpStore: "copy_OB_to_OS",
	IT_CopyOpStoreToOpA: "copy_OS_to_OA",
	IT_CopyOpStoreToOpB: "copy_OS_to_OB",
	IT_CopyOpAToReturn:  "copy_OA_to_RET",
	IT_CopyReturnToOpA:  "copy_RET_to_OA",
	IT_CopyReturnToOpB:  "copy_RET_to_OB",
	IT_CopyOpAToListA:   "copy_OA_to_LA",
	IT_CopyListAToOpA:   "copy_LA_to_OA",
	IT_CopyListAToOpB:   "copy_LA_to_OB",

	IT_PushOpAToArg:    "push_OA_to_ARG",
	IT_PushOpBToArg:    "push_OB_to_ARG",
	IT_PopArgStackRegA: "pop_ARG_to_OA",

	IT_IntAdd:      "int_add",
	IT_IntSubtract: "int_sub",
	IT_IntMultiply: "int_mul",
	IT_IntDivide:   "int_div",
	IT_IntPower:    "int_pow",
	IT_IntModulo:   "int_mod",

	IT_FloatAdd:      "flt_add",
	IT_FloatSubtract: "flt_sub",
	IT_FloatMultiply: "flt_mul",
	IT_FloatDivide:   "flt_div",
	IT_FloatPower:    "flt_pow",
	IT_FloatModulo:   "flt_mod",

	IT_StringConcat: "str_concat",

	IT_DeclareBool:   "decl_bool",
	IT_DeclareInt:    "decl_int",
	IT_DeclareFloat:  "decl_float",
	IT_DeclareString: "decl_string",
	IT_DeclareList:   "decl_list",

	IT_Return: "return",

	IT_Equal:             "equal",
	IT_IntLower:          "int_lower",
	IT_IntGreater:        "int_greater",
	IT_IntLowerEqual:     "int_lower_equal",
	IT_IntGreaterEqual:   "int_greater_equal",
	IT_FloatLower:        "float_lower",
	IT_FloatGreater:      "float_greater",
	IT_FloatLowerEqual:   "float_lower_equal",
	IT_FloatGreaterEqual: "float_greater_equal",
	IT_Not:               "not",

	IT_SetRegAFalse: "set_OA_false",
	IT_SetRegATrue:  "set_OA_true",
	IT_SetRegBFalse: "set_OB_false",
	IT_SetRegBTrue:  "set_OB_true",

	IT_PushScopeUnnamed: "push_scope_unnamed",
	IT_PopScope:         "pop_scope",

	IT_CreateListInListA: "new_list_LA",
	IT_AppendOpAToListA:  "append_OA_to_LA",

	IT_LineOffset: "line_offset",
}

type Instruction struct {
	InstructionType  byte
	InstructionValue []byte
}

type ExpandedInstruction struct {
	InstructionType  byte
	InstructionValue []int
}
