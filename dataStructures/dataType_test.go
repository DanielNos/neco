package dataStructures

import (
	"testing"
)

func TestIsCompositeType(t *testing.T) {
	dataTypes := map[*DataType]bool{
		{DT_Unknown, nil}:                     false,
		{DT_Any, nil}:                         false,
		{DT_None, nil}:                        false,
		{DT_Bool, nil}:                        false,
		{DT_Int, nil}:                         false,
		{DT_Float, nil}:                       false,
		{DT_String, nil}:                      false,
		{DT_Enum, nil}:                        false,
		{DT_Object, nil}:                      false,
		{DT_List, &DataType{DT_Bool, nil}}:    true,
		{DT_Set, &DataType{DT_String, nil}}:   true,
		{DT_Option, &DataType{DT_Float, nil}}: true,
	}

	for dataType, isComposite := range dataTypes {
		if dataType.IsCompositeType() != isComposite {
			t.Errorf("(%s).IsCompositeType(): %v, want %v", dataType, dataType.IsCompositeType(), isComposite)
		}
	}
}

func TestEquals(t *testing.T) {
	intType := &DataType{DT_Int, nil}
	noneType := &DataType{DT_None, nil}
	strOptionType := &DataType{DT_Option, &DataType{DT_String, nil}}
	enumtype := &DataType{DT_Enum, "Day"}
	objectType := &DataType{DT_Object, "Person"}
	boolListType := &DataType{DT_List, &DataType{DT_Bool, nil}}

	// Equals
	allTypes := []*DataType{intType, noneType, strOptionType, enumtype, objectType, boolListType}
	for _, dataType := range allTypes {
		if !dataType.Equals(dataType) {
			t.Errorf("%s == %s: false, want true", dataType, dataType)
		}
	}

	// Not equals
	combinations := map[*DataType]*DataType{
		intType:       noneType,
		boolListType:  intType,
		enumtype:      intType,
		enumtype:      strOptionType,
		enumtype:      objectType,
		strOptionType: boolListType,
	}

	for type1, type2 := range combinations {
		if type1.Equals(type2) {
			t.Errorf("%s == %s: true, want false", type1, type2)
		}
	}
}

func TestAssign(t *testing.T) {
	unknownType := &DataType{DT_Unknown, nil}
	anyType := &DataType{DT_Any, nil}
	intType := &DataType{DT_Int, nil}
	boolType := &DataType{DT_Bool, nil}
	noneType := &DataType{DT_None, nil}
	intOptionType := &DataType{DT_Option, &DataType{DT_Int, nil}}
	enumtype := &DataType{DT_Enum, "Day"}
	objectType1 := &DataType{DT_Object, "Person"}
	objectType2 := &DataType{DT_Object, "Animal"}
	intListType := &DataType{DT_List, &DataType{DT_Int, nil}}

	// Assignable
	assignableCombinations := map[*DataType]*DataType{
		anyType:       anyType,
		anyType:       intType,
		anyType:       noneType,
		anyType:       objectType2,
		anyType:       intListType,
		intType:       intType,
		noneType:      noneType,
		intOptionType: intOptionType,
		intOptionType: intType,
		intOptionType: noneType,
		enumtype:      enumtype,
		objectType1:   objectType1,
		intListType:   intListType,
		intListType:   intListType,
	}

	for assignedToType, assignedType := range assignableCombinations {
		if !assignedToType.CanBeAssigned(assignedType) {
			t.Errorf("(%s).CanBeAssigned(%s): false, want true", assignedToType, assignedType)
		}
	}

	// Not assignable
	notAssignableCombinations := map[*DataType]*DataType{
		unknownType:   anyType,
		anyType:       unknownType,
		intType:       unknownType,
		unknownType:   intType,
		intType:       anyType,
		intOptionType: anyType,
		noneType:      anyType,
		intType:       boolType,
		intOptionType: boolType,
		intListType:   intType,
		intListType:   intOptionType,
		objectType1:   objectType2,
		noneType:      intOptionType,
	}

	for assignedToType, assignedType := range notAssignableCombinations {
		if assignedToType.CanBeAssigned(assignedType) {
			t.Errorf("(%s).CanBeAssigned(%s): true, want false", assignedToType, assignedType)
		}
	}
}

func TestIsComplete(t *testing.T) {
	// Complete types
	completeTypes := []*DataType{
		{DT_Any, nil},
		{DT_None, nil},
		{DT_Int, nil},
		{DT_Object, "Person"},
		{DT_Enum, "Day"},
		{DT_Option, &DataType{DT_Int, nil}},
		{DT_Set, &DataType{DT_Float, nil}},
		{DT_List, &DataType{DT_Bool, nil}},
		{DT_Option, &DataType{DT_List, &DataType{DT_String, nil}}},
		{DT_Set, &DataType{DT_List, &DataType{DT_Option, &DataType{DT_Bool, nil}}}},
	}

	for _, dataType := range completeTypes {
		if !dataType.IsComplete() {
			t.Errorf("(%s).IsComplete(): false, want true", dataType)
		}
	}

	// Incomplete types
	incompleteTypes := []*DataType{
		{DT_Unknown, nil},
		{DT_Option, &DataType{DT_Unknown, nil}},
		{DT_Set, &DataType{DT_Unknown, nil}},
		{DT_List, &DataType{DT_Unknown, nil}},
		{DT_Option, &DataType{DT_List, &DataType{DT_Unknown, nil}}},
		{DT_Set, &DataType{DT_List, &DataType{DT_Option, &DataType{DT_Unknown, nil}}}},
	}

	for _, dataType := range incompleteTypes {
		if dataType.IsComplete() {
			t.Errorf("(%s).IsComplete(): true, want false", dataType)
		}
	}
}

func TestCopy(t *testing.T) {
	dataTypes := []*DataType{
		{DT_Unknown, nil},
		{DT_None, nil},
		{DT_Any, nil},
		{DT_Bool, nil},
		{DT_Int, nil},
		{DT_Float, nil},
		{DT_String, nil},
		{DT_Enum, "Day"},
		{Type: DT_Object, SubType: "Person"},
		{DT_List, &DataType{DT_Bool, nil}},
		{DT_Set, &DataType{DT_Int, nil}},
		{DT_Option, &DataType{DT_Float, nil}},
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}},
	}

	for _, dataType := range dataTypes {
		copiedType := dataType.Copy()

		if !dataType.Equals(copiedType) {
			t.Errorf("(%s).Copy(): %s, want %s", dataType, copiedType, dataType)
		}
	}
}

func TryCompleteFrom(t *testing.T) {
	intType := &DataType{DT_Int, nil}
	listFloatType := &DataType{DT_List, &DataType{DT_Float, nil}}
	optionBoolType := &DataType{DT_Option, &DataType{DT_Bool, nil}}
	enumType := &DataType{DT_Enum, "Day"}
	setObjectType := &DataType{DT_Set, &DataType{DT_Object, "Person"}}

	unknownType := &DataType{DT_Unknown, nil}
	listUnknownType := &DataType{DT_List, &DataType{DT_Unknown, nil}}
	optionUnknownType := &DataType{DT_Option, &DataType{DT_Unknown, nil}}
	setUnkownType := &DataType{DT_Set, &DataType{DT_Unknown, nil}}

	typesToComplete := []*DataType{
		intType,
		unknownType,
		unknownType,
		listFloatType,
		listFloatType,
		enumType,
		listUnknownType,
		listUnknownType,
		listUnknownType,
		listUnknownType,
		optionBoolType,
		optionUnknownType,
		optionUnknownType,
		optionUnknownType,
		setObjectType,
		setUnkownType,
		setUnkownType,
	}

	typesToCompleteFrom := []*DataType{
		intType,
		intType,
		unknownType,
		intType,
		optionBoolType,
		intType,
		listUnknownType,
		unknownType,
		intType,
		listFloatType,
		optionUnknownType,
		listFloatType,
		enumType,
		optionBoolType,
		intType,
		intType,
		setObjectType,
	}

	completedTypes := []*DataType{
		intType,
		intType,
		unknownType,
		listFloatType,
		listFloatType,
		enumType,
		listUnknownType,
		listUnknownType,
		listUnknownType,
		listFloatType,
		optionBoolType,
		optionUnknownType,
		optionUnknownType,
		optionBoolType,
		setObjectType,
		setUnkownType,
		setObjectType,
	}

	for i := 0; i < len(completedTypes); i++ {
		originalTypeName := typesToComplete[i].String()
		typesToComplete[i].TryCompleteFrom(typesToCompleteFrom[i])

		if !typesToComplete[i].Equals(completedTypes[i]) {
			t.Errorf("(%s).Equals(%s): %s, want %s", originalTypeName, typesToCompleteFrom[i], typesToComplete[i], completedTypes[i])
		}
	}
}

func TestGetDepth(t *testing.T) {
	dataTypes := map[*DataType]int{
		{DT_Unknown, nil}:                     1,
		{DT_None, nil}:                        1,
		{DT_Any, nil}:                         1,
		{DT_Bool, nil}:                        1,
		{DT_Int, nil}:                         1,
		{DT_Float, nil}:                       1,
		{DT_String, nil}:                      1,
		{DT_Enum, "Day"}:                      1,
		{DT_Object, "Person"}:                 1,
		{DT_List, &DataType{DT_Bool, nil}}:    2,
		{DT_Set, &DataType{DT_Int, nil}}:      2,
		{DT_Option, &DataType{DT_Float, nil}}: 2,
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}}:    3,
		{DT_Set, &DataType{DT_List, &DataType{DT_Unknown, nil}}}: 3,
	}

	for dataType, depth := range dataTypes {
		if dataType.GetDepth() != depth {
			t.Errorf("(%s).GetDepth(): %d, want %d", dataType, dataType.GetDepth(), depth)
		}
	}
}

func TestGetLeafType(t *testing.T) {
	dataTypes := map[*DataType]*DataType{
		{DT_Unknown, nil}:                     {DT_Unknown, nil},
		{DT_None, nil}:                        {DT_None, nil},
		{DT_Any, nil}:                         {DT_Any, nil},
		{DT_Bool, nil}:                        {DT_Bool, nil},
		{DT_Int, nil}:                         {DT_Int, nil},
		{DT_Float, nil}:                       {DT_Float, nil},
		{DT_String, nil}:                      {DT_String, nil},
		{DT_Enum, "Day"}:                      {DT_Enum, "Day"},
		{DT_Object, "Person"}:                 {DT_Object, "Person"},
		{DT_List, &DataType{DT_Bool, nil}}:    {DT_Bool, nil},
		{DT_Set, &DataType{DT_Int, nil}}:      {DT_Int, nil},
		{DT_Option, &DataType{DT_Float, nil}}: {DT_Float, nil},
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}}:    {DT_Bool, nil},
		{DT_Set, &DataType{DT_List, &DataType{DT_Unknown, nil}}}: {DT_Unknown, nil},
	}

	for dataType, correctLeafType := range dataTypes {
		leafType := dataType.GetLeafType()

		if !leafType.Equals(correctLeafType) {
			t.Errorf("(%s).GetLeafType(): %s, want %s", dataType, leafType, correctLeafType)
		}
	}
}

func TestSetLeafType(t *testing.T) {
	dataTypes := []*DataType{
		{DT_Unknown, nil},
		{DT_None, nil},
		{DT_Any, nil},
		{DT_Bool, nil},
		{DT_Int, nil},
		{DT_Float, nil},
		{DT_String, nil},
		{DT_Enum, "Day"},
		{DT_Object, "Person"},
		{DT_List, &DataType{DT_Bool, nil}},
		{DT_Set, &DataType{DT_Int, nil}},
		{DT_Option, &DataType{DT_Float, nil}},
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}},
		{DT_Set, &DataType{DT_List, &DataType{DT_Unknown, nil}}},
	}

	for _, dataType := range dataTypes {
		typeName := dataType.String()
		dataType.SetLeafType(&DataType{DT_String, nil})

		if dataType.GetLeafType().Type != DT_String {
			t.Errorf("(%s).SetLeafType(string): %s, want string", typeName, dataType.GetLeafType())
		}
	}
}

func TestDataTypeString(t *testing.T) {
	dataTypes := map[*DataType]string{
		{DT_Unknown, nil}:                     "?",
		{DT_None, nil}:                        "none",
		{DT_Any, nil}:                         "any",
		{DT_Bool, nil}:                        "bool",
		{DT_Int, nil}:                         "int",
		{DT_Float, nil}:                       "float",
		{DT_String, nil}:                      "string",
		{DT_Enum, "Day"}:                      "Day",
		{DT_Object, "Person"}:                 "Person",
		{DT_List, &DataType{DT_Bool, nil}}:    "list<bool>",
		{DT_Set, &DataType{DT_Int, nil}}:      "set<int>",
		{DT_Option, &DataType{DT_Float, nil}}: "float?",
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}}:    "set<list<bool>>",
		{DT_Set, &DataType{DT_List, &DataType{DT_Unknown, nil}}}: "set<list<?>>",
		{DT_List, &DataType{DT_Option, &DataType{DT_Int, nil}}}:  "list<int?>",
	}

	for dataType, name := range dataTypes {
		if dataType.String() != name {
			t.Errorf("(%s).String(): %s, want %s", dataType, dataType, name)
		}
	}
}

func TestSignature(t *testing.T) {
	dataTypes := map[*DataType]string{
		{DT_Unknown, nil}:                     "?",
		{DT_None, nil}:                        "none",
		{DT_Any, nil}:                         "any",
		{DT_Bool, nil}:                        "bool",
		{DT_Int, nil}:                         "int",
		{DT_Float, nil}:                       "float",
		{DT_String, nil}:                      "string",
		{DT_Enum, "Day"}:                      "enum:Day",
		{DT_Object, "Person"}:                 "Person",
		{DT_List, &DataType{DT_Bool, nil}}:    "list<bool>",
		{DT_Set, &DataType{DT_Int, nil}}:      "set<int>",
		{DT_Option, &DataType{DT_Float, nil}}: "opt<float>",
		{DT_Set, &DataType{DT_List, &DataType{DT_Bool, nil}}}:    "set<list<bool>>",
		{DT_Set, &DataType{DT_List, &DataType{DT_Unknown, nil}}}: "set<list<?>>",
		{DT_List, &DataType{DT_Option, &DataType{DT_Int, nil}}}:  "list<int?>",
	}

	for dataType, name := range dataTypes {
		if dataType.Signature() != name {
			t.Errorf("(%s).Signature(): %s, want %s", dataType, dataType.Signature(), name)
		}
	}
}
