package parser

type Symbol struct {
	symbolType SymbolType
	value SymbolValue
}

type SymbolType uint8

const (
	ST_Variable SymbolType = iota
	ST_Function
	ST_Struct
	ST_Enum
)

type SymbolValue interface{}

type VariableSymbol struct {
	dataType DataType
	canBeNone bool
	isDeclared bool
}
