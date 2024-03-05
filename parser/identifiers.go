package parser

import (
	"fmt"
	data "neco/dataStructures"
	"neco/lexer"
)

func (p *Parser) parseIdentifier() *Node {
	identifier := p.consume()
	symbol := p.findSymbol(identifier.Value)

	// Decalare enum
	if symbol != nil && symbol.symbolType == ST_Enum {
		p.consume()
		return &Node{}
	}

	// Assign to variable
	if p.peek().TokenType.IsAssignKeyword() {
		var expression *Node
		// Undeclared symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			expression, _ = p.parseAssign([]*lexer.Token{identifier}, []data.DataType{{data.DT_NoType, nil}})
		} else {
			// Assignment to function
			if symbol.symbolType == ST_Function {
				p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []data.DataType{{data.DT_NoType, nil}})
				// Assignment to variable
			} else {
				// Can't assign to constants
				if symbol.value.(*VariableSymbol).isConstant {
					p.newError(p.peek().Position, fmt.Sprintf("Can't assign to constant variable %s.", identifier.Value))
				}

				expression, _ = p.parseAssign([]*lexer.Token{identifier}, []data.DataType{symbol.value.(*VariableSymbol).VariableType})

				symbol.value.(*VariableSymbol).isInitialized = true
			}
		}
		return expression
	}

	// Assign to list at index
	if p.peek().TokenType == lexer.TT_DL_BracketOpen {
		var listType data.DataType = data.DataType{data.DT_NoType, nil}

		// Undeclared symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
		} else {
			// Not a variable
			if symbol.symbolType != ST_Variable {
				p.newError(identifier.Position, fmt.Sprintf("Can't assign to %s. It is not a variable.", identifier.Value))
				// Collect list type
			} else {
				listType = symbol.value.(*VariableSymbol).VariableType
			}
		}

		// Collect index expression
		p.consume()
		indexExpression := p.parseExpressionRoot()
		p.consume()

		// Index must be int
		if !p.GetExpressionType(indexExpression).Equals(data.DataType{data.DT_Int, nil}) {
			p.newError(indexExpression.Position, "Index expression has to be int.")
		}

		assignPosition := p.consume().Position

		// Collect assigned expression
		assignedExpression := p.parseExpressionRoot()
		assignedType := p.GetExpressionType(assignedExpression)

		// Check if assigned expression has the correct type
		if listType.DType != data.DT_NoType && !assignedType.Equals(listType.SubType.(data.DataType)) {
			p.newError(assignedExpression.Position, fmt.Sprintf("Can't assign expression of type %s to %s.", assignedType, listType))
		}

		return &Node{assignPosition, NT_ListAssign, &ListAssignNode{identifier.Value, symbol.value.(*VariableSymbol), indexExpression, assignedExpression}}
	}

	// Assign to multiple variables
	if p.peek().TokenType == lexer.TT_DL_Comma {
		var dataTypes = []data.DataType{}

		// Check symbol
		if symbol == nil {
			p.newError(identifier.Position, fmt.Sprintf("Use of undeclared variable %s.", identifier.Value))
			dataTypes = append(dataTypes, data.DataType{data.DT_NoType, nil})
		} else if symbol.symbolType == ST_Function {
			p.newError(identifier.Position, fmt.Sprintf("Can't assign to function %s.", identifier.Value))
			dataTypes = append(dataTypes, data.DataType{data.DT_NoType, nil})
		} else {
			dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).VariableType)
		}

		// Collect identifiers
		var identifiers = []*lexer.Token{identifier}
		symbols := []*Symbol{}

		for p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()

			// Look up identifier and collect it
			symbol = p.findSymbol(p.peek().Value)
			identifiers = append(identifiers, p.consume())

			// Check symbol
			if symbol == nil {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Use of undeclared variable %s.", identifiers[len(identifiers)-1]))
				dataTypes = append(dataTypes, data.DataType{data.DT_NoType, nil})
			} else if symbol.symbolType == ST_Function {
				p.newError(p.peekPrevious().Position, fmt.Sprintf("Can't assign to function %s.", identifiers[len(identifiers)-1]))
				dataTypes = append(dataTypes, data.DataType{data.DT_NoType, nil})
			} else {
				dataTypes = append(dataTypes, symbol.value.(*VariableSymbol).VariableType)
			}

			symbols = append(symbols, symbol)
		}

		expression, _ := p.parseAssign(identifiers, dataTypes)

		// Set symbols as initialized
		for _, symbol := range symbols {
			symbol.value.(*VariableSymbol).isInitialized = true
		}

		return expression
	}

	// Function call
	// Undeclared function
	if symbol == nil {
		p.newError(identifier.Position, fmt.Sprintf("Use of undeclared function %s.", identifier.Value))
		return p.parseFunctionCall(symbol, identifier)
	}

	// Declared function
	return p.parseFunctionCall(symbol, identifier)
}
