package main

import "fmt"

type TokenType int8

const (
	TT_EndOfCommand TokenType = iota
	TT_EndOfFile

	TT_Identifier
	TT_Operator
	TT_Delimiter

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
)

var TokenTypeToString = map[TokenType]string {
	TT_EndOfCommand: "EndOfCommand",
	TT_EndOfFile: "EndOfFile",
	
	TT_Identifier: "Identifier",
	TT_Operator: "Operator",
	TT_Delimiter: "Delimiter",

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
	return fmt.Sprintf("%s %d:%d\t  %v\t  %s", *t.position.file, t.position.startLine, t.position.startChar, t.tokenType, t.value)
}
