package dataStructures

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

func (dt DataType) CanBeAssigned(other DataType) bool {
	// No type can't equal any other type
	if dt.Type == DT_NoType || other.Type == DT_NoType {
		return false
	}

	// Compare primitive data types
	if dt.Type <= DT_String && other.Type <= DT_String {
		return dt.Type == other.Type
	}

	// Any can be assigned anything (except NoType)
	if dt.Type == DT_Any {
		return true
	}

	// Lists
	if dt.Type == DT_List && other.Type == DT_List {
		return dt.SubType.(DataType).CanBeAssigned(other.SubType.(DataType))
	}

	// Compare struct names
	if dt.Type == DT_Struct && other.Type == DT_Struct {
		return dt.SubType == other.SubType
	}

	// Compare enum names
	if dt.Type == DT_Enum && other.Type == DT_Enum {
		if dt.SubType == nil || other.SubType == nil {
			return true
		}
		return dt.SubType == other.SubType
	}

	// Compare sets
	if dt.Type == DT_Set && other.Type == DT_Set {
		return dt.SubType.(DataType).CanBeAssigned(other.SubType.(DataType))
	}

	return false
}

func (dt DataType) String() string {
	if dt.Type <= DT_Any {
		return dt.Type.String()
	} else if dt.Type <= DT_Struct {
		if dt.SubType == nil {
			return "Any"
		}
		return dt.SubType.(string)
	} else {
		if dt.SubType == nil {
			return dt.Type.String() + "<?>"
		}
		return dt.Type.String() + "<" + dt.SubType.(DataType).String() + ">"
	}
}

func (dt DataType) Signature() string {
	if dt.Type <= DT_Any {
		return dt.Type.String()
	} else if dt.Type == DT_Enum {
		return "Enum:" + dt.SubType.(string)
	} else if dt.Type == DT_Struct {
		if dt.SubType == nil {
			return "Any"
		}
		return dt.SubType.(string)
	} else {
		return dt.Type.String() + "<" + dt.SubType.(DataType).String() + ">"
	}
}
