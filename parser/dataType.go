package parser

import (
	"fmt"
	"neco/lexer"
)

type DType uint8

const (
	DT_NoType DType = iota
	DT_Bool
	DT_Int
	DT_Float
	DT_String
	DT_Any
	DT_UserDefined
	DT_List
)

func (dt DType) String() string {
	switch dt {
	case DT_NoType:
		return "No Type"
	case DT_Bool:
		return "Bool"
	case DT_Int:
		return "Int"
	case DT_Float:
		return "Float"
	case DT_String:
		return "String"
	case DT_Any:
		return "Any"
	case DT_List:
		return "List"
	case DT_UserDefined:
		return "Custom"
	}

	return "[INVALID DATA TYPE]"
}

var TokenTypeToDataType = map[lexer.TokenType]DType{
	lexer.TT_KW_var: DT_NoType,

	lexer.TT_KW_bool: DT_Bool,
	lexer.TT_LT_Bool: DT_Bool,

	lexer.TT_KW_int: DT_Int,
	lexer.TT_LT_Int: DT_Int,

	lexer.TT_KW_flt:   DT_Float,
	lexer.TT_LT_Float: DT_Float,

	lexer.TT_KW_str:    DT_String,
	lexer.TT_LT_String: DT_String,

	lexer.TT_KW_list: DT_List,
}

type DataType struct {
	DType   DType
	SubType interface{}
}

func (vt DataType) Equals(other DataType) bool {
	return vt.DType != DT_NoType && other.DType != DT_Any && (vt.DType == DT_Any || other.DType == DT_Any) || vt.DType == other.DType && vt.SubType == other.SubType
}

func (dt DataType) String() string {
	if dt.DType <= DT_UserDefined {
		return dt.DType.String()
	}

	return fmt.Sprintf("%s<%s>", dt.DType, dt.SubType.(DataType))
}
