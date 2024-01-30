package parser

func (p *Parser) insertBuiltInFunctions() {
	p.symbolTableStack.Top.Value.(symbolTable)["print"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_String, false}, "text", nil}}, VariableType{DT_NoType, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["printLine"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_String, false}, "text", nil}}, VariableType{DT_NoType, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["boolToStr"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_String, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["intToStr"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_String, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["fltToStr"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_String, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["boolToInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Bool, false}, "boolean", nil}}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["intToFlt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Float, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["floor"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["ceil"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["round"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["floorToInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["ceilToInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["roundToInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Int, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["absInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Int, false}, "integer", nil}}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["absFlt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Float, false}, "float", nil}}, VariableType{DT_Float, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["readLine"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{}, VariableType{DT_String, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["readChar"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{}, VariableType{DT_String, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["length"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["toLower"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["toUpper"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_String, false}, "string", nil}}, VariableType{DT_String, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["randomInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{}, VariableType{DT_Int, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["randomFlt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{}, VariableType{DT_Float, false}}}
	p.symbolTableStack.Top.Value.(symbolTable)["randomRangeInt"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Int, false}, "min", nil}, {VariableType{DT_Int, false}, "max", nil}}, VariableType{DT_Int, false}}}

	p.symbolTableStack.Top.Value.(symbolTable)["exit"] = &Symbol{ST_Function, &FunctionSymbol{[]Parameter{{VariableType{DT_Int, false}, "exitCode", nil}}, VariableType{DT_NoType, false}}}
}
