package parser

import data "neco/dataStructures"

var NO_PARAMS = []Parameter{}

func (p *Parser) insertBuiltInFunctions() {
	p.insertFunction("print", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_String, nil}, "text", nil}},
		data.DataType{data.DT_NoType, nil}, true},
	)

	p.insertFunction("printLine", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_String, nil}, "text", nil}},
		data.DataType{data.DT_NoType, nil}, true},
	)

	p.insertFunction("str", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Any, nil}, "value", nil}},
		data.DataType{data.DT_String, nil}, true},
	)

	p.insertFunction("int", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Bool, nil}, "boolean", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("flt", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Int, nil}, "integer", nil}},
		data.DataType{data.DT_Float, nil}, true},
	)

	p.insertFunction("floor", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Float, nil}, true},
	)
	p.insertFunction("ceil", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Float, nil}, true},
	)
	p.insertFunction("round", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Float, nil}, true},
	)

	p.insertFunction("floorToInt", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("ceilToInt", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("roundata.DToInt", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)

	p.insertFunction("abs", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Int, nil}, "integer", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("abs", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Float, nil}, "float", nil}},
		data.DataType{data.DT_Float, nil}, true},
	)

	p.insertFunction("readLine", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_String, nil}, true})
	p.insertFunction("readChar", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_String, nil}, true})

	p.insertFunction("length", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_String, nil}, "string", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)
	p.insertFunction("length", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_List, data.DataType{data.DT_Any, nil}}, "list", nil}},
		data.DataType{data.DT_Int, nil}, true},
	)

	p.insertFunction("toLower", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_String, nil}, "string", nil}},
		data.DataType{data.DT_String, nil}, true},
	)
	p.insertFunction("toUpper", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_String, nil}, "string", nil}},
		data.DataType{data.DT_String, nil}, true},
	)

	p.insertFunction("randomInt", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_Int, nil}, true})
	p.insertFunction("randomFlt", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_Float, nil}, true})
	p.insertFunction("randomRangeInt", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Int, nil}, "min", nil}, {data.DataType{data.DT_Int, nil}, "max", nil}},
		data.DataType{data.DT_Int, false}, true},
	)

	p.insertFunction("exit", &FunctionSymbol{-1,
		[]Parameter{{data.DataType{data.DT_Int, nil}, "exitCode", nil}},
		data.DataType{data.DT_NoType, nil}, true},
	)

	p.insertFunction("line", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_NoType, nil}, true})

	// Debug fucntions
	p.insertFunction("trace", &FunctionSymbol{-1, NO_PARAMS, data.DataType{data.DT_NoType, nil}, true})
}
