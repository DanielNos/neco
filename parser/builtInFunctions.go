package parser

func (p *Parser) insertBuiltInFunctions() {
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "bool", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_NoType, false}, true})

	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "bool", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_NoType, false}, true})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_NoType, false}, true})

	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_String, false}, true})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_String, false}, true})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_String, false}, true})

	p.insertFunction("int", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_Int, false}, true})
	p.insertFunction("flt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Float, false}, true})

	p.insertFunction("floor", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}, true})
	p.insertFunction("ceil", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}, true})
	p.insertFunction("round", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}, true})

	p.insertFunction("floorToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}, true})
	p.insertFunction("ceilToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}, true})
	p.insertFunction("roundToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}, true})

	p.insertFunction("abs", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Int, false}, true})
	p.insertFunction("abs", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}, true})

	p.insertFunction("readLine", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_String, false}, true})
	p.insertFunction("readChar", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_String, false}, true})

	p.insertFunction("length", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_Int, false}, true})
	p.insertFunction("toLower", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}, true})
	p.insertFunction("toUpper", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}, true})

	p.insertFunction("randomInt", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_Int, false}, true})
	p.insertFunction("randomFlt", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_Float, false}, true})
	p.insertFunction("randomRangeInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "min", nil}, {VariableType{DT_Int, false}, "max", nil}}, VariableType{DT_Int, false}, true})

	p.insertFunction("exit", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "exitCode", nil}}, VariableType{DT_NoType, false}, true})

	p.insertFunction("line", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_NoType, false}, true})
}
