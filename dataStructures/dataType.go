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
	DT_Object
	DT_List
	DT_Set
)

func (pt PrimitiveType) String() string {
	switch pt {
	case DT_NoType:
		return "[NOTYPE]"
	case DT_Bool:
		return "bool"
	case DT_Int:
		return "int"
	case DT_Float:
		return "float"
	case DT_String:
		return "string"
	case DT_Any:
		return "any"
	case DT_Enum:
		return "enum"
	case DT_Object:
		return "object"
	case DT_List:
		return "list"
	case DT_Set:
		return "set"
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
	if dt.Type == DT_Object && other.Type == DT_Object {
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
	} else if dt.Type <= DT_Object {
		if dt.SubType == nil {
			return "any"
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
		return "enum:" + dt.SubType.(string)
	} else if dt.Type == DT_Object {
		if dt.SubType == nil {
			return "any"
		}
		return dt.SubType.(string)
	} else {
		return dt.Type.String() + "<" + dt.SubType.(DataType).String() + ">"
	}
}
