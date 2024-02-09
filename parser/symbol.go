package parser

import "fmt"

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

type SymbolValue interface{}

type VariableSymbol struct {
	variableType  VariableType
	isInitialized bool
	isConstant    bool
}

type FunctionSymbol struct {
	number     int
	parameters []Parameter
	returnType VariableType
	everCalled bool
}

type symbolTable map[string]*Symbol

func (p *Parser) findSymbol(identifier string) *Symbol {
	stackNode := p.symbolTableStack.Top

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
	symbol, exists := p.symbolTableStack.Top.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}

	return nil
}

func (p *Parser) getGlobalSymbol(identifier string) *Symbol {
	symbol, exists := p.symbolTableStack.Bottom.Value.(symbolTable)[identifier]

	if exists {
		return symbol
	}
	return nil
}

func (p *Parser) insertFunction(name string, functionSymbol *FunctionSymbol) *Symbol {
	// Find bucket
	bucket, exists := p.symbolTableStack.Bottom.Value.(symbolTable)[name]

	// Create bucket if it doesn't exist
	if !exists {
		bucket = &Symbol{ST_FunctionBucket, symbolTable{}}
		p.symbolTableStack.Bottom.Value.(symbolTable)[name] = bucket
	}

	// Insert function in to bucket
	symbol := &Symbol{ST_Function, functionSymbol}
	bucket.value.(symbolTable)[createParametersIdentifier(functionSymbol.parameters)] = symbol

	return symbol
}

func createParametersIdentifier(parameters []Parameter) string {
	id := ""
	for _, parameter := range parameters {
		id = fmt.Sprintf("%s.%s", id, parameter.DataType)
	}

	return id
}
