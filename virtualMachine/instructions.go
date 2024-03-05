package virtualMachine

const (
	// 1 argument 2 bytes
	IT_JumpEx byte = iota
	IT_JumpIfTrueEx
	IT_JumpBackEx

	// 1 argument 1 byte
	IT_Call
	IT_CallBuiltInFunc
	IT_PushScope

	IT_Halt

	IT_DeclareBool
	IT_DeclareInt
	IT_DeclareFloat
	IT_DeclareString
	IT_DeclareList

	IT_SetListAtPrevToCurr

	IT_LoadConst
	IT_LoadConstToList
	IT_Load
	IT_Store

	IT_JumpBack
	IT_Jump
	IT_JumpIfTrue

	// 0 arguments
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

	IT_And
	IT_Or

	IT_StringConcat
	IT_ListConcat

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

	IT_PushTrue
	IT_PushFalse

	IT_PushScopeUnnamed
	IT_PopScope

	IT_CreateList
	IT_AppendToList
	IT_IndexList

	IT_LineOffset
)

var InstructionTypeToString = map[byte]string{
	// 1 argument 2 bytes
	IT_JumpEx:       "jump_ex",
	IT_JumpIfTrueEx: "jump_if_true_ex",
	IT_JumpBackEx:   "jump_back_ex",

	// 1 argument 1 byte
	IT_Call:            "call",
	IT_CallBuiltInFunc: "call_builtin",
	IT_PushScope:       "push_scope",

	IT_Halt: "halt",

	IT_DeclareBool:   "decl_bool",
	IT_DeclareInt:    "decl_int",
	IT_DeclareFloat:  "decl_float",
	IT_DeclareString: "decl_string",
	IT_DeclareList:   "decl_list",

	IT_SetListAtPrevToCurr: "set_list",

	IT_LoadConst:       "load_const",
	IT_LoadConstToList: "append_const",
	IT_Load:            "load",
	IT_Store:           "store",

	IT_JumpBack:   "jump_back",
	IT_Jump:       "jump",
	IT_JumpIfTrue: "jump_if_true",

	// 0 arguments
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

	IT_And: "and",
	IT_Or:  "or",

	IT_StringConcat: "string_concat",
	IT_ListConcat:   "list_concat",

	IT_Return: "return",

	IT_Equal:             "equal",
	IT_IntLower:          "int_lower",
	IT_FloatLower:        "flt_lower",
	IT_IntGreater:        "int_greater",
	IT_FloatGreater:      "flt_greater",
	IT_IntLowerEqual:     "int_lower_equal",
	IT_FloatLowerEqual:   "flt_lower_equal",
	IT_IntGreaterEqual:   "int_greater_equal",
	IT_FloatGreaterEqual: "flt_greater_equal",
	IT_Not:               "not",

	IT_PushTrue:  "push_true",
	IT_PushFalse: "push_false",

	IT_PushScopeUnnamed: "push_scope_unnamed",
	IT_PopScope:         "pop_scope",

	IT_CreateList:   "new_list",
	IT_AppendToList: "append_list",
	IT_IndexList:    "index_list",

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

func IsJumpForward(instructionType byte) bool {
	return instructionType == IT_Jump || instructionType == IT_JumpEx || instructionType == IT_JumpIfTrue || instructionType == IT_JumpIfTrueEx
}
