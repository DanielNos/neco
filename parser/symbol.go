package parser

import (
	data "neco/dataStructures"
)

type Symbol struct {
	symbolType SymbolType
	value      SymbolValue
}

type SymbolType uint8

const (
	ST_Variable SymbolType = iota
	ST_FunctionBucket
	ST_Function
	ST_Struct
	ST_Enum
)

func (st SymbolType) String() string {
	switch st {
	case ST_Variable:
		return "variable"
	case ST_FunctionBucket, ST_Function:
		return "function"
	case ST_Struct:
		return "struct"
	case ST_Enum:
		return "enum"
	}

	return "UNDEFINED"
}

type SymbolValue interface{}

type VariableSymbol struct {
	VariableType  data.DataType
	isInitialized bool
	isConstant    bool
}

type FunctionSymbol struct {
	number     int
	parameters []Parameter
	returnType data.DataType
	everCalled bool
}

type PropertySymbol struct {
	number   int
	dataType data.DataType
}

type symbolTable map[string]*Symbol

func (p *Parser) insertSymbol(key string, symbol *Symbol) {
	p.stack_symbolTableStack.Top.Value.(symbolTable)[key] = symbol
}

func (p *Parser) findSymbol(identifier string) *Symbol {
	stackNode := p.stack_symbolTableStack.Top

	for stackNode != nil {
		symbol, exists := stackNode.Value.(symbolTable)[identifier]

		if exists {
			return symbol
		}

		stackNode = stackNode.Previous
	}

	return nil
}

func (p *Parser) getSymbol(identifier string) *Symbol {
	symbol, exists := p.stack_symbolTableStack.Top.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}

	return nil
}

func (p *Parser) getGlobalSymbol(identifier string) *Symbol {
	symbol, exists := p.stack_symbolTableStack.Bottom.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}
	return nil
}

func (p *Parser) insertFunction(name string, functionSymbol *FunctionSymbol) *Symbol {
	// Find bucket
	bucket, exists := p.stack_symbolTableStack.Bottom.Value.(symbolTable)[name]

	// Create bucket if it doesn't exist
	if !exists {
		bucket = &Symbol{ST_FunctionBucket, symbolTable{}}
		p.stack_symbolTableStack.Bottom.Value.(symbolTable)[name] = bucket
	}

	// Insert function in to bucket
	symbol := &Symbol{ST_Function, functionSymbol}
	bucket.value.(symbolTable)[createParametersIdentifier(functionSymbol.parameters)] = symbol

	return symbol
}

func createParametersIdentifier(parameters []Parameter) string {
	id := ""
	for _, parameter := range parameters {
		id += "." + parameter.DataType.String()
	}

	return id
}
