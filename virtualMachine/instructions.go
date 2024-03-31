package virtualMachine

import "fmt"

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
	IT_DeclareSet
	IT_DeclareObject
	IT_DeclareOption

	IT_SetListAtPrevToCurr

	IT_LoadConst
	IT_LoadConstToList
	IT_Load
	IT_Store
	IT_StoreAndPop

	IT_CreateObject
	IT_GetField

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
	IT_PushNone

	IT_PushScopeUnnamed
	IT_PopScope

	IT_StoreField

	IT_CreateList
	IT_AppendToList
	IT_IndexList

	IT_CreateSet
	IT_InsertToSet
	IT_SetContains
	IT_RemoveSetElement

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
	IT_DeclareSet:    "decl_set",
	IT_DeclareObject: "decl_object",
	IT_DeclareOption: "decl_option",

	IT_SetListAtPrevToCurr: "set_list",

	IT_LoadConst:       "load_const",
	IT_LoadConstToList: "append_const",
	IT_Load:            "load",
	IT_Store:           "store",
	IT_StoreAndPop:     "store_pop",

	IT_CreateObject: "new_object",
	IT_GetField:     "get_field",

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
	IT_IntLower:          "int_lwr",
	IT_FloatLower:        "flt_lwr",
	IT_IntGreater:        "int_gtr",
	IT_FloatGreater:      "flt_gtr",
	IT_IntLowerEqual:     "int_lwr_eq",
	IT_FloatLowerEqual:   "flt_lwr_eq",
	IT_IntGreaterEqual:   "int_gtr_eq",
	IT_FloatGreaterEqual: "flt_gtr_eq",
	IT_Not:               "not",

	IT_PushTrue:  "push_true",
	IT_PushFalse: "push_false",
	IT_PushNone:  "push_none",

	IT_PushScopeUnnamed: "push_scope_unnamed",
	IT_PopScope:         "pop_scope",

	IT_StoreField: "store_field",

	IT_CreateList:   "list_new",
	IT_AppendToList: "list_append",
	IT_IndexList:    "list_index",

	IT_CreateSet:        "set_new",
	IT_InsertToSet:      "set_insert",
	IT_SetContains:      "set_contains",
	IT_RemoveSetElement: "set_remove",

	IT_LineOffset: "line_offset",
}

type Instruction struct {
	InstructionType  byte
	InstructionValue []byte
}

func (i Instruction) String() string {
	if len(i.InstructionValue) == 0 {
		return InstructionTypeToString[i.InstructionType] + ";"
	}

	return InstructionTypeToString[i.InstructionType] + " " + fmt.Sprintf("%d", i.InstructionValue[0]) + ";"
}

type ExpandedInstruction struct {
	InstructionType  byte
	InstructionValue []int
}

func IsJumpForward(instructionType byte) bool {
	return instructionType == IT_Jump || instructionType == IT_JumpEx || instructionType == IT_JumpIfTrue || instructionType == IT_JumpIfTrueEx
}

func IsCompositeDeclarator(instructionType byte) bool {
	return instructionType == IT_DeclareList || instructionType == IT_DeclareSet
}
