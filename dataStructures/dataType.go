package dataStructures

import "fmt"

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

type DataType struct {
	DType   DType
	SubType interface{}
}

func (vt DataType) Equals(other DataType) bool {
	if vt.DType == DT_NoType || other.DType == DT_NoType {
		return false
	}

	if vt.DType <= DT_String && other.DType <= DT_String {
		return vt.DType == other.DType
	}

	if vt.DType == DT_Any {
		return true
	}

	if vt.DType == DT_List && other.DType == DT_List {
		return vt.SubType.(DataType).Equals(other.SubType.(DataType))
	}

	return false
}

func (dt DataType) String() string {
	if dt.DType <= DT_UserDefined {
		return dt.DType.String()
	}

	return fmt.Sprintf("%s<%s>", dt.DType, dt.SubType.(DataType))
}
