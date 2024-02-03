package parser

func (p *Parser) insertBuiltInFunctions() {
	p.insertFunction("print", []Parameter{{VariableType{DT_String, false}, "text", nil}}, VariableType{DT_NoType, false})
	p.insertFunction("printLine", []Parameter{{VariableType{DT_String, false}, "text", nil}}, VariableType{DT_NoType, false})

	p.insertFunction("boolToStr", []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_String, false})
	p.insertFunction("intToStr", []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_String, false})
	p.insertFunction("fltToStr", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_String, false})

	p.insertFunction("boolToInt", []Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_Int, false})
	p.insertFunction("intToFlt", []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Float, false})

	p.insertFunction("floor", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false})
	p.insertFunction("ceil", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false})
	p.insertFunction("round", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false})

	p.insertFunction("floorToInt", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false})
	p.insertFunction("ceilToInt", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false})
	p.insertFunction("roundToInt", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false})

	p.insertFunction("absInt", []Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Int, false})
	p.insertFunction("absFlt", []Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false})

	p.insertFunction("readLine", []Parameter{}, VariableType{DT_String, false})
	p.insertFunction("readChar", []Parameter{}, VariableType{DT_String, false})

	p.insertFunction("length", []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_Int, false})
	p.insertFunction("toLower", []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false})
	p.insertFunction("toUpper", []Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false})

	p.insertFunction("randomInt", []Parameter{}, VariableType{DT_Int, false})
	p.insertFunction("randomFlt", []Parameter{}, VariableType{DT_Float, false})
	p.insertFunction("randomRangeInt", []Parameter{{VariableType{DT_Int, false}, "min", nil}, {VariableType{DT_Int, false}, "max", nil}}, VariableType{DT_Int, false})

	p.insertFunction("exit", []Parameter{{VariableType{DT_Int, false}, "exitCode", nil}}, VariableType{DT_NoType, false})

	p.insertFunction("line", []Parameter{}, VariableType{DT_NoType, false})
}
