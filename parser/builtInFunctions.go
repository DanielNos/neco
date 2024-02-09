package parser

func (p *Parser) insertBuiltInFunctions() {
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "bool", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("print", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_NoType, false}})

	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "bool", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_NoType, false}})
	p.insertFunction("printLine", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_NoType, false}})

	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_String, false}})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_String, false}})
	p.insertFunction("str", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_String, false}})

	p.insertFunction("int", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_Int, false}})
	p.insertFunction("flt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Float, false}})

	p.insertFunction("floor", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}})
	p.insertFunction("ceil", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}})
	p.insertFunction("round", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}})

	p.insertFunction("floorToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}})
	p.insertFunction("ceilToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}})
	p.insertFunction("roundToInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}})

	p.insertFunction("absInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Int, false}})
	p.insertFunction("absFlt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}})

	p.insertFunction("readLine", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_String, false}})
	p.insertFunction("readChar", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_String, false}})

	p.insertFunction("length", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_Int, false}})
	p.insertFunction("toLower", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}})
	p.insertFunction("toUpper", &FunctionSymbol{-1, []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}})

	p.insertFunction("randomInt", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_Int, false}})
	p.insertFunction("randomFlt", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_Float, false}})
	p.insertFunction("randomRangeInt", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "min", nil}, {VariableType{DT_Int, false}, "max", nil}}, VariableType{DT_Int, false}})

	p.insertFunction("exit", &FunctionSymbol{-1, []Parameter{{VariableType{DT_Int, false}, "exitCode", nil}}, VariableType{DT_NoType, false}})

	p.insertFunction("line", &FunctionSymbol{-1, []Parameter{}, VariableType{DT_NoType, false}})
}
