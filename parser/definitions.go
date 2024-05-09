package parser

import (
	data "neco/dataStructures"
	"neco/lexer"
	"neco/logger"
)

func (p *Parser) collectGlobals() {
	// Collect enums
	for p.peek().TokenType != lexer.TT_EndOfFile {
		if p.peek().TokenType == lexer.TT_KW_enum {
			p.parseEnum()
		} else {
			p.consume()
		}
	}
	p.tokenIndex = 1

	// Collect struct names
	for p.peek().TokenType != lexer.TT_EndOfFile {
		if p.peek().TokenType == lexer.TT_KW_struct {
			p.consume()

			symbol := p.getGlobalSymbol(p.peek().Value)

			if symbol != nil {
				p.newError(p.peek().Position, "Symbol is already declared as a "+symbol.symbolType.String()+".")
			}

			p.insertSymbol(p.consume().Value, &Symbol{ST_Struct, nil})
		} else {
			p.consume()
		}
	}
	p.tokenIndex = 1

	// Collect structs
	for p.peek().TokenType != lexer.TT_EndOfFile {
		if p.peek().TokenType == lexer.TT_KW_struct {
			p.parseStruct()
		} else {
			p.consume()
		}
	}
	p.tokenIndex = 1

	// Collect function headers
	for p.peek().TokenType != lexer.TT_EndOfFile {
		if p.peek().TokenType == lexer.TT_KW_fun {
			p.consume()
			p.parseFunctionHeader()
		} else {
			p.consume()
		}
	}
	p.tokenIndex = 1

	// Collect global variables
	scopeDepth := 0
	for p.peek().TokenType != lexer.TT_EndOfFile {
		// Enter scope
		if p.peek().TokenType == lexer.TT_DL_BraceOpen {
			scopeDepth++
			// Leave scope
		} else if p.peek().TokenType == lexer.TT_DL_BraceClose {
			scopeDepth--
			// Collect globals only in root scope
		} else if scopeDepth == 0 && p.peek().TokenType == lexer.TT_EndOfCommand {
			p.consume()
			if p.peek().TokenType.IsVariableType() {
				p.appendScope(p.parseVariableDeclaration(false))
				continue
			} else if p.peek().TokenType == lexer.TT_KW_const {
				p.consume()
				p.appendScope(p.parseVariableDeclaration(true))
				continue
			}
		}
		p.consume()
	}
	p.tokenIndex = 1
}

func (p *Parser) parseStruct() {
	p.consume()

	// Collect symbol
	identifier := p.consume()
	symbol := p.getGlobalSymbol(identifier.Value)

	p.consume() // {
	p.consumeEOCs()

	// Collect properties
	properties := map[string]PropertySymbol{}
	propertyIndex := 0

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Collect property
		dataType := p.parseType()
		properties[p.consume().Value] = PropertySymbol{propertyIndex, dataType}
		propertyIndex++

		// Collect properties with same type
		for p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()

			properties[p.consume().Value] = PropertySymbol{propertyIndex, dataType}
			propertyIndex++
		}

		p.consume()
	}

	p.consume() // }

	if len(properties) == 0 {
		if p.ErrorCount+p.totalErrorCount == 0 {
			println()
		}
		logger.WarningCodePos(identifier.Position, "Struct "+identifier.Value+" has no fields.")
	}

	symbol.value = properties
}

func (p *Parser) parseEnum() {
	p.consume()
	// Collect identifier
	identifier := p.consume().Value

	if p.peek().TokenType == lexer.TT_EndOfCommand {
		p.consume()
	}

	// Check for redeclaration
	symbol := p.getGlobalSymbol(identifier)

	if symbol != nil {
		p.newError(p.peekPrevious().Position, "Symbol is already declared as a "+symbol.symbolType.String()+".")
	}

	p.consume() // {

	// Consume EOCs
	p.consumeEOCs()

	// Collect enum constants
	constants := map[string]int64{}
	constantIndex := int64(0)

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Collect identifier
		constantIdentifier := p.consume()

		// Change index
		if p.peek().TokenType == lexer.TT_KW_Assign {
			p.consume()

			expression := p.parseExpressionRoot()

			// New index can't be lower than current index
			if expression.Value.(*LiteralNode).Value.(int64) < constantIndex {
				p.newError(expression.Position, "Constant indexes can't be used for multiple enumerator constants.")
			}

			constantIndex = expression.Value.(*LiteralNode).Value.(int64)
		}

		// Check if constant identifier already exists
		if _, exists := constants[constantIdentifier.Value]; exists {
			p.newError(constantIdentifier.Position, "Duplicate enum constant identifier.")
		}

		// Store constant
		constants[constantIdentifier.Value] = int64(constantIndex)
		constantIndex++

		// Consume EOCs
		p.consumeEOCs()
	}

	p.consume() // }

	p.insertSymbol(identifier, &Symbol{ST_Enum, constants})
}

func (p *Parser) parseFunctionHeader() {
	// Find bucket
	identifierToken := p.consume()
	symbol := p.findSymbol(identifierToken.Value)

	// Enter scope
	p.enterScope()
	p.consume() // (

	// Collect parameters
	startPosition := p.peek().Position
	parameters := p.parseParameters()

	// Function entry() can't have parameters
	if identifierToken.Value == "entry" && len(parameters) != 0 {
		p.newError(startPosition.Combine(p.peekPrevious().Position), "Function entry() can't have any parameters.")
	}

	// Check for redeclaration
	if symbol != nil {
		// Redeclaration of entry()
		if identifierToken.Value == "entry" {
			p.newError(identifierToken.Position, "Function entry() can't be overloaded.")
		}

		// Create parameters id and look for a function using it
		if symbol.symbolType == ST_FunctionBucket {
			id := createParametersIdentifier(parameters)
			if symbol.value.(symbolTable)[id] != nil {
				p.newError(identifierToken.Position, "Redeclaration of symbol "+identifierToken.Value+".")
			}
		}
	}

	p.consume() // )

	// Collect return type
	returnType := &data.DataType{data.DT_Unknown, nil}
	var returnPosition *data.CodePos

	if p.peek().TokenType == lexer.TT_KW_returns {
		returnPosition = p.consume().Position
		returnType.Type = TokenTypeToDataType[p.consume().TokenType]
		returnPosition.EndChar = p.peekPrevious().Position.EndChar

		// Function entry() can't have a return type
		if identifierToken.Value == "entry" {
			p.newError(returnPosition, "Function entry() can't have a return type.")
		}
	}

	// Leave scope
	p.scopeNodeStack.Pop()
	p.stack_symbolTableStack.Pop()

	// Insert function symbol
	newSymbol := p.insertFunction(identifierToken.Value, &FunctionSymbol{len(p.functions), parameters, returnType, identifierToken.Value == "entry"})
	p.functions = append(p.functions, newSymbol.value.(*FunctionSymbol))
}
