package virtualMachine

import "neco/parser"

type SymbolType uint8

const (
	ST_Variable SymbolType = iota
)

type Symbol struct {
	symbolType  SymbolType
	symbolValue interface{}
}

type VariableSymbol struct {
	dataType parser.DType
	value    interface{}
}
