package main

import "fmt"

type TokenType int8

const (
	TT_EndOfCommand TokenType = iota
	TT_EndOfFile

	TT_Identifier

	TT_DL_ParenthesisOpen
	TT_DL_ParenthesisClose
	TT_DL_BracketOpen
	TT_DL_BracketClose
	TT_DL_BraceOpen
	TT_DL_BraceClose
	TT_DL_Comma

	TT_OP_Add
	TT_OP_Subtract
	TT_OP_Multiply
	TT_OP_Divide
	TT_OP_Modulo
	TT_OP_Equal
	TT_OP_Not
	TT_OP_NotEqual
	TT_OP_Lower
	TT_OP_LowerEqual
	TT_OP_Greater
	TT_OP_GreaterEqual

	TT_LT_Bool
	TT_LT_Int
	TT_LT_Float
	TT_LT_String

	TT_KW_const
	TT_KW_var
	TT_KW_bool
	TT_KW_int
	TT_KW_flt
	TT_KW_str

	TT_KW_Assign
	TT_KW_AddAssign
	TT_KW_SubtractAssign
	TT_KW_MultiplyAssign
	TT_KW_DivideAssign
	TT_KW_ModuloAssign

	TT_KW_loop
	TT_KW_while
	TT_KW_for
	TT_KW_forEach
)

var TokenTypeToString = map[TokenType]string {
	TT_EndOfCommand: "EndOfCommand",
	TT_EndOfFile: "EndOfFile",
	
	TT_Identifier: "Identifier",

	TT_DL_ParenthesisOpen: "(",
	TT_DL_ParenthesisClose: ")",
	TT_DL_BracketOpen: "[",
	TT_DL_BracketClose: "]",
	TT_DL_BraceOpen: "{",
	TT_DL_BraceClose: "}",

	TT_OP_Add: "+",
	TT_OP_Subtract: "-",
	TT_OP_Multiply: "*",
	TT_OP_Divide: "/",
	TT_OP_Modulo: "%",
	TT_OP_Equal: "==",
	TT_OP_Not: "!",
	TT_OP_NotEqual: "!=",
	TT_OP_Lower: "<",
	TT_OP_LowerEqual: "<=",
	TT_OP_Greater: ">",
	TT_OP_GreaterEqual: ">=",

	TT_LT_Bool: "Bool",
	TT_LT_Int: "Int",
	TT_LT_Float: "Float",
	TT_LT_String: "String",

	TT_KW_const: "const",
	TT_KW_var: "var",
	TT_KW_bool: "bool",
	TT_KW_int: "int",
	TT_KW_flt: "flt",
	TT_KW_str: "str",

	TT_KW_Assign: "=",
	TT_KW_AddAssign: "+=",
	TT_KW_SubtractAssign: "-=",
	TT_KW_MultiplyAssign: "*=",
	TT_KW_DivideAssign: "/=",
	TT_KW_ModuloAssign: "%=",

	TT_KW_loop: "loop",
	TT_KW_while: "while",
	TT_KW_for: "for",
	TT_KW_forEach: "forEach",
}

func (tt TokenType) String() string {
	return TokenTypeToString[tt]
}

type Token struct {
	position *CodePos
	tokenType TokenType
	value string
}

func (t *Token) String() string {
	label := fmt.Sprintf("%v", t.tokenType)

	for len(label) < 14 {
		label = fmt.Sprintf("%s ", label)
	}

	message := fmt.Sprintf("%s %d:%d\t  %s %s", *t.position.file, t.position.startLine, t.position.startChar, label, t.value)

	return message
}
