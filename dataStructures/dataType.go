package dataStructures

type PrimitiveType uint8

const (
	DT_Unknown PrimitiveType = iota
	DT_Bool
	DT_Int
	DT_Float
	DT_String
	DT_Any
	DT_None
	DT_Enum
	DT_Object
	DT_List
	DT_Set
	DT_Option
)

func (pt PrimitiveType) String() string {
	switch pt {
	case DT_Unknown:
		return "?"
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
	case DT_None:
		return "none"
	case DT_Enum:
		return "enum"
	case DT_Object:
		return "object"
	case DT_List:
		return "list"
	case DT_Set:
		return "set"
	case DT_Option:
		return "opt"
	}

	return "[INVALID DATA TYPE]"
}

type DataType struct {
	Type    PrimitiveType
	SubType any
}

func (dt *DataType) CanBeAssigned(other *DataType) bool {
	// No type can't equal any other type
	if dt.Type == DT_Unknown || other.Type == DT_Unknown {
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
		return dt.SubType.(*DataType).CanBeAssigned(other.SubType.(*DataType))
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
		return dt.SubType.(*DataType).CanBeAssigned(other.SubType.(*DataType))
	}

	// Compare options/nones
	if dt.Type == DT_Option {
		if other.Type == DT_None {
			return true
		}

		if other.IsCompositeType() {
			return dt.SubType.(*DataType).CanBeAssigned(other.SubType.(*DataType))
		}

		return dt.SubType.(*DataType).CanBeAssigned(other)
	}

	// Nones
	if dt.Type == DT_None && other.Type == DT_None {
		return true
	}

	return false
}

func (dt *DataType) Equals(other *DataType) bool {
	if dt.Type != other.Type {
		return false
	}

	if dt.Type <= DT_None {
		return true
		// User defined types
	} else if dt.Type <= DT_Object {
		return dt.SubType == other.SubType
		// Composite types
	} else {
		return dt.SubType.(*DataType).Equals(other.SubType.(*DataType))
	}
}

func (dt *DataType) IsComplete() bool {
	if dt.SubType == nil || dt.Type == DT_Enum || dt.Type == DT_Object {
		return dt.Type != DT_Unknown
	}

	return dt.SubType.(*DataType).IsComplete()
}

func (dt *DataType) Copy() *DataType {
	if dt.Type <= DT_Object {
		return &DataType{dt.Type, dt.SubType}
	}

	return &DataType{dt.Type, dt.SubType.(*DataType).Copy()}
}

func (dt *DataType) TryCompleteFrom(from *DataType) {
	// Data type can't be completed
	if dt.GetDepth() > from.GetDepth() {
		return
	}

	if dt.IsCompositeType() && dt.SubType != nil {
		dt.SubType.(*DataType).TryCompleteFrom(from.SubType.(*DataType))
	} else {
		dt.Type = from.Type
		dt.SubType = from.SubType
	}
}

func (dt *DataType) GetDepth() int {
	if dt.IsCompositeType() {
		return 1 + dt.SubType.(*DataType).GetDepth()
	}

	return 1
}

func (dt *DataType) GetLeafType() *DataType {
	if dt.IsCompositeType() {
		return dt.SubType.(*DataType).GetLeafType()
	}

	return dt
}

func (dt *DataType) SetLeafType(dataType *DataType) {
	if dt.IsCompositeType() {
		dt.SubType.(*DataType).SetLeafType(dataType)
	} else {
		*dt = *dataType
	}
}

func (dt *DataType) IsCompositeType() bool {
	return dt.Type >= DT_List && dt.Type <= DT_Option
}

func (dt *DataType) String() string {
	// Literals
	if dt.Type <= DT_None {
		return dt.Type.String()
		// User defined types
	} else if dt.Type <= DT_Object {
		if dt.SubType == nil {
			return "any"
		}
		return dt.SubType.(string)
		// Composite types
	} else if dt.Type <= DT_Set {
		if dt.SubType == nil {
			return dt.Type.String() + "<?>"
		}
		return dt.Type.String() + "<" + dt.SubType.(*DataType).String() + ">"
		// Option type
	} else {
		return dt.SubType.(*DataType).String() + "?"
	}
}

func (dt *DataType) Signature() string {
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
		return dt.Type.String() + "<" + dt.SubType.(*DataType).String() + ">"
	}
}
