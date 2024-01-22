package parser

import (
	"fmt"
	"neco/lexer"
)

type DataType uint8

const (
	DT_NoType DataType = iota
	DT_None
	DT_Bool
	DT_Int
	DT_Float
	DT_String
)

func (dt DataType) String() string {
	switch dt {
	case DT_NoType:
		return "No Type"
	case DT_None:
		return "None"
	case DT_Bool:
		return "Bool"
	case DT_Int:
		return "Int"
	case DT_Float:
		return "Float"
	case DT_String:
		return "String"
	}

	return "[INVALID DATA TYPE]"
}

var TokenTypeToDataType = map[lexer.TokenType]DataType{
	lexer.TT_KW_var: DT_NoType,

	lexer.TT_LT_None: DT_None,

	lexer.TT_KW_bool: DT_Bool,
	lexer.TT_LT_Bool: DT_Bool,

	lexer.TT_KW_int: DT_Int,
	lexer.TT_LT_Int: DT_Int,

	lexer.TT_KW_flt:   DT_Float,
	lexer.TT_LT_Float: DT_Float,

	lexer.TT_KW_str:    DT_String,
	lexer.TT_LT_String: DT_String,
}

type VariableType struct {
	DataType  DataType
	CanBeNone bool
}

func (vt VariableType) Equals(other VariableType) bool {
	return vt.DataType == other.DataType && vt.CanBeNone == other.CanBeNone
}

func (v VariableType) String() string {
	if v.CanBeNone {
		return fmt.Sprintf("%s?", v.DataType)
	}
	return v.DataType.String()
}
