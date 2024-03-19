package dataStructures

import (
	"fmt"
)

type DType uint8

const (
	DT_NoType DType = iota
	DT_Bool
	DT_Int
	DT_Float
	DT_String
	DT_Any
	DT_Enum
	DT_Struct
	DT_List
	DT_Set
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
	case DT_Enum:
		return "Enum"
	case DT_Struct:
		return "Struct"
	case DT_List:
		return "List"
	case DT_Set:
		return "Set"
	}

	return "[INVALID DATA TYPE]"
}

type DataType struct {
	DType   DType
	SubType interface{}
}

func (vt DataType) Equals(other DataType) bool {
	// No type can't equal any other type
	if vt.DType == DT_NoType || other.DType == DT_NoType {
		return false
	}

	// Compare basic data types
	if vt.DType <= DT_String && other.DType <= DT_String {
		return vt.DType == other.DType
	}

	// Any can be assigned anything (except NoType)
	if vt.DType == DT_Any {
		return true
	}

	// Lists
	if vt.DType == DT_List && other.DType == DT_List {
		return vt.SubType.(DataType).Equals(other.SubType.(DataType))
	}

	// Compare struct names
	if vt.DType == DT_Struct && other.DType == DT_Struct {
		return vt.SubType == other.SubType
	}

	// Compare enum names
	if vt.DType == DT_Enum && other.DType == DT_Enum {
		if vt.SubType == nil || other.SubType == nil {
			return true
		}
		return vt.SubType == other.SubType
	}

	// Compare structs
	if vt.DType == DT_Set && other.DType == DT_Set {
		return vt.SubType == other.SubType
	}

	return false
}

func (dt DataType) String() string {
	if dt.DType <= DT_Any {
		return dt.DType.String()
	} else if dt.DType <= DT_Struct {
		return dt.SubType.(string)
	} else {
		return fmt.Sprintf("%s<%s>", dt.DType, dt.SubType.(DataType))
	}
}

func (dt DataType) Signature() string {
	if dt.DType <= DT_Any {
		return dt.DType.String()
	} else if dt.DType == DT_Enum {
		return "Enum"
	} else if dt.DType == DT_Struct {
		return dt.SubType.(string)
	} else {
		return fmt.Sprintf("%s<%s>", dt.DType, dt.SubType.(DataType))
	}
}
