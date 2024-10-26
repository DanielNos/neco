package parser

import data "github.com/DanielNos/neco/dataStructures"

var NO_PARAMS = []Parameter{}

func (p *Parser) insertBuiltInFunctions() {
	// Prints
	p.insertFunction("print", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "text", nil}},
		nil, true},
	)
	p.insertFunction("printLine", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "text", nil}},
		nil, true},
	)
	p.insertFunction("printLine", &FunctionSymbol{number: -1, parameters: []Parameter{}, returnType: nil, everCalled: true})

	// To string
	p.insertFunction("str", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Any, nil}, "value", nil}},
		&data.DataType{data.DT_String, nil}, true},
	)

	// To int
	p.insertFunction("int", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Bool, nil}, "boolean", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("int", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Enum, nil}, "enum", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)

	// To float
	p.insertFunction("flt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Int, nil}, "integer", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)

	// Round/ceil/floor float
	p.insertFunction("floor", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)
	p.insertFunction("ceil", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)
	p.insertFunction("round", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)

	// Round/ceil/floor to int
	p.insertFunction("floorToInt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("ceilToInt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("roundToInt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)

	// Absolute values
	p.insertFunction("abs", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Int, nil}, "integer", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("abs", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Float, nil}, "float", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)

	// Reading text from terminal
	p.insertFunction("readLine", &FunctionSymbol{-1, NO_PARAMS, &data.DataType{data.DT_String, nil}, true})
	p.insertFunction("readChar", &FunctionSymbol{-1, NO_PARAMS, &data.DataType{data.DT_String, nil}, true})

	// Length of strings
	p.insertFunction("length", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "string", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)

	// Size of lists
	p.insertFunction("size", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_List, &data.DataType{data.DT_Any, nil}}, "list", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)

	// String to upper/lower
	p.insertFunction("toLower", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "string", nil}},
		&data.DataType{data.DT_String, nil}, true},
	)
	p.insertFunction("toUpper", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "string", nil}},
		&data.DataType{data.DT_String, nil}, true},
	)

	// Random ints/floats
	p.insertFunction("randomInt", &FunctionSymbol{-1, NO_PARAMS, &data.DataType{data.DT_Int, nil}, true})
	p.insertFunction("randomFlt", &FunctionSymbol{-1, NO_PARAMS, &data.DataType{data.DT_Float, nil}, true})
	p.insertFunction("randomRangeInt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Int, nil}, "min", nil}, {&data.DataType{data.DT_Int, nil}, "max", nil}},
		&data.DataType{data.DT_Int, false}, true},
	)

	// Exit
	p.insertFunction("exit", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_Int, nil}, "exitCode", nil}},
		nil, true},
	)

	// Parsing numbers
	p.insertFunction("parseInt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "string", nil}},
		&data.DataType{data.DT_Int, nil}, true},
	)

	p.insertFunction("parseFlt", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "string", nil}},
		&data.DataType{data.DT_Float, nil}, true},
	)

	// Debug functions
	p.insertFunction("trace", &FunctionSymbol{-1, NO_PARAMS, nil, true})

	p.insertFunction("panic", &FunctionSymbol{-1,
		[]Parameter{{&data.DataType{data.DT_String, nil}, "message", nil}},
		nil, true},
	)
}
