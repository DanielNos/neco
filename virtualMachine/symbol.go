package virtualMachine

import data "neco/dataStructures"

type SymbolType uint8

const (
	ST_Variable SymbolType = iota
)

type Symbol struct {
	symbolType  SymbolType
	symbolValue interface{}
}

type VariableSymbol struct {
	dataType data.DataType
	value    interface{}
}
