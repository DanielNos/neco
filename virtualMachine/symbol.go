package virtualMachine

import data "github.com/DanielNos/neco/dataStructures"

type SymbolType uint8

const (
	ST_Variable SymbolType = iota
)

type Symbol struct {
	symbolType  SymbolType
	symbolValue any
}

type VariableSymbol struct {
	dataType data.DataType
	value    any
}
