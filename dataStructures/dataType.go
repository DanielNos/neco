package dataStructures

import (
	"fmt"
)

type PrimitiveType uint8

const (
	DT_NoType PrimitiveType = iota
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

func (pt PrimitiveType) String() string {
	switch pt {
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
	Type    PrimitiveType
	SubType interface{}
}

func (vt DataType) CanBeAssigned(other DataType) bool {
	// No type can't equal any other type
	if vt.Type == DT_NoType || other.Type == DT_NoType {
		return false
	}

	// Compare primitive data types
	if vt.Type <= DT_String && other.Type <= DT_String {
		return vt.Type == other.Type
	}

	// Any can be assigned anything (except NoType)
	if vt.Type == DT_Any {
		return true
	}

	// Lists
	if vt.Type == DT_List && other.Type == DT_List {
		return vt.SubType.(DataType).CanBeAssigned(other.SubType.(DataType))
	}

	// Compare struct names
	if vt.Type == DT_Struct && other.Type == DT_Struct {
		return vt.SubType == other.SubType
	}

	// Compare enum names
	if vt.Type == DT_Enum && other.Type == DT_Enum {
		if vt.SubType == nil || other.SubType == nil {
			return true
		}
		return vt.SubType == other.SubType
	}

	// Compare sets
	if vt.Type == DT_Set && other.Type == DT_Set {
		return vt.SubType.(DataType).CanBeAssigned(other.SubType.(DataType))
	}

	return false
}

func (dt DataType) String() string {
	if dt.Type <= DT_Any {
		return dt.Type.String()
	} else if dt.Type <= DT_Struct {
		return dt.SubType.(string)
	} else {
		return fmt.Sprintf("%s<%s>", dt.Type, dt.SubType.(DataType))
	}
}

func (dt DataType) Signature() string {
	if dt.Type <= DT_Any {
		return dt.Type.String()
	} else if dt.Type == DT_Enum {
		return "Enum:" + dt.SubType.(string)
	} else if dt.Type == DT_Struct {
		return dt.SubType.(string)
	} else {
		return fmt.Sprintf("%s<%s>", dt.Type, dt.SubType.(DataType))
	}
}
