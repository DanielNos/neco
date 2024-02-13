package parser

var NO_PARAMS = []Parameter{}

func (p *Parser) insertBuiltInFunctions() {
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, nil}, "bool", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "integer", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, nil}, "string", nil}}, VariableType{DT_NoType, nil}, true})

	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, nil}, "bool", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "integer", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_NoType, nil}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, nil}, "string", nil}}, VariableType{DT_NoType, nil}, true})

	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, nil}, "boolean", nil}}, VariableType{DT_String, nil}, true})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "integer", nil}}, VariableType{DT_String, nil}, true})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_String, nil}, true})

	p.insertFunction("int", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, nil}, "boolean", nil}}, VariableType{DT_Int, nil}, true})
	p.insertFunction("flt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "integer", nil}}, VariableType{DT_Float, nil}, true})

	p.insertFunction("floor", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Float, nil}, true})
	p.insertFunction("ceil", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Float, nil}, true})
	p.insertFunction("round", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Float, nil}, true})

	p.insertFunction("floorToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Int, nil}, true})
	p.insertFunction("ceilToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Int, nil}, true})
	p.insertFunction("roundToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Int, nil}, true})

	p.insertFunction("abs", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "integer", nil}}, VariableType{DT_Int, nil}, true})
	p.insertFunction("abs", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, nil}, "float", nil}}, VariableType{DT_Float, nil}, true})

	p.insertFunction("readLine", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_String, nil}, true})
	p.insertFunction("readChar", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_String, nil}, true})

	p.insertFunction("length", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, nil}, "string", nil}}, VariableType{DT_Int, nil}, true})
	p.insertFunction("toLower", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, nil}, "string", nil}}, VariableType{DT_String, nil}, true})
	p.insertFunction("toUpper", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, nil}, "string", nil}}, VariableType{DT_String, nil}, true})

	p.insertFunction("randomInt", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_Int, nil}, true})
	p.insertFunction("randomFlt", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_Float, nil}, true})
	p.insertFunction("randomRangeInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "min", nil}, {VariableType{DT_Int, nil}, "max", nil}}, VariableType{DT_Int, false}, true})

	p.insertFunction("exit", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, nil}, "exitCode", nil}}, VariableType{DT_NoType, nil}, true})

	p.insertFunction("line", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_NoType, nil}, true})

	// Debug fucntions
	p.insertFunction("trace", &FunctionSymbol{-1, NO_PARAMS, VariableType{DT_NoType, nil}, true})
}
