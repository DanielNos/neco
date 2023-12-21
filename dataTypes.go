package main

type DataType uint8

const (
	DT_None DataType = iota
	DT_Bool
	DT_Int
	DT_Float
	DT_String
)

func (d DataType) String() string {
	switch d {
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

var TokenTypeToDataType = map[TokenType]DataType {
	TT_LT_None: DT_None,

	TT_KW_bool: DT_Bool,
	TT_LT_Bool: DT_Bool,

	TT_KW_int: DT_Int,
	TT_LT_Int: DT_Int,

	TT_KW_flt: DT_Float,
	TT_LT_Float: DT_Float,

	TT_KW_str: DT_String,
	TT_LT_String: DT_String,
}
