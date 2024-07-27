package parser

import (
	"fmt"
	"math"
	"strconv"

	data "github.com/DanielNos/neco/dataStructures"
	"github.com/DanielNos/neco/lexer"
	VM "github.com/DanielNos/neco/virtualMachine"
)

func (p *Parser) parseLiteral() *Node {
	var literalValue LiteralValue

	switch p.peek().TokenType {
	case lexer.TT_LT_Bool:
		literalValue = p.peek().Value[0] == '1'
	case lexer.TT_LT_Int:
		literalValue, _ = strconv.ParseInt(p.peek().Value, 10, 64)
	case lexer.TT_LT_Float:
		literalValue, _ = strconv.ParseFloat(p.peek().Value, 64)
	case lexer.TT_LT_String:
		literalValue = p.peek().Value
	case lexer.TT_LT_None:
		literalValue = nil
	}

	return &Node{p.peek().Position, NT_Literal, &LiteralNode{TokenTypeToDataType[p.consume().TokenType], literalValue}}
}

func (p *Parser) parseStructLiteral(properties map[string]PropertySymbol) *Node {
	identifier := p.consume()
	p.StringConstants[identifier.Value] = -1

	p.consume() // {
	p.consumeEOCs()

	var propertyValues []*Node

	// Collect named properties
	if p.peek().TokenType == lexer.TT_Identifier && p.peekNext().TokenType == lexer.TT_DL_Colon {
		propertyValues = p.parseKeyedProperties(properties, identifier.Value)
	} else {
		propertyValues = p.parseProperties(properties, identifier)
	}

	p.consume() // }

	return &Node{identifier.Position.Combine(p.peekPrevious().Position), NT_Object, &ObjectNode{identifier.Value, propertyValues}}
}

func (p *Parser) parseKeyedProperties(properties map[string]PropertySymbol, structName string) []*Node {
	propertyValues := map[string]*Node{}

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Field doesn't have a key
		if p.peek().TokenType != lexer.TT_Identifier || p.peekNext().TokenType != lexer.TT_DL_Colon {

			p.newError(p.peek().Position, "All values have to have keys in keyed struct creation.")

			// Collect expression
			p.parseExpressionRoot()

		} else {
			// Collect key
			propertyName := p.consume()
			p.consume()

			// Look up property
			property, exists := properties[propertyName.Value]

			// It doesn't exist
			if !exists {
				p.newError(propertyName.Position, "Struct "+structName+" doesn't have a field "+propertyName.Value+".")
				// It exists
			} else {
				// Check if property is already assigned
				_, isReassigned := propertyValues[propertyName.Value]

				// It's already assigned
				if isReassigned {
					p.newError(propertyName.Position, "Field "+propertyName.Value+" is already assigned.")
				}
			}

			// Collect expression
			expression := p.parseExpressionRoot()

			// Check if expression has correct type
			if exists {
				expressionType := GetExpressionType(expression)
				if !property.dataType.CanBeAssigned(expressionType) {
					p.newError(expression.Position, "Field "+propertyName.Value+" of struct "+structName+" has type "+property.dataType.String()+", but is assigned expression of type "+expressionType.String()+".")
				}
			}

			// Store field value
			propertyValues[propertyName.Value] = expression
		}

		// More fields
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
			p.consumeEOCs()
		} else {
			for p.peek().TokenType != lexer.TT_DL_BraceClose {
				p.newError(p.peek().Position, "Unexpected token after struct field value.")
			}
		}
	}

	// Change order of values to match property order
	orderedValues := make([]*Node, len(properties))

	for key, property := range properties {
		propertyValue, exists := propertyValues[key]

		if exists {
			orderedValues[property.number] = propertyValue
		}
	}

	return orderedValues
}

func (p *Parser) parseProperties(properties map[string]PropertySymbol, structName *lexer.Token) []*Node {
	// Make properties linear
	orderedProperties := make([]PropertySymbol, len(properties))
	orderedPropertyNames := make([]string, len(properties))

	for key, property := range properties {
		orderedPropertyNames[property.number] = key
		orderedProperties[property.number] = property
	}

	// Collect field values
	propertyValues := make([]*Node, len(properties))
	propertyIndex := 0

	for p.peek().TokenType != lexer.TT_DL_BraceClose {
		// Too many fields
		if propertyIndex == len(properties) {
			p.newError(p.peek().Position, "Struct "+structName.Value+fmt.Sprintf(" has %d fields, but %d values were provided.", len(properties), propertyIndex+1))
			p.parseExpressionRoot()
			// Collect field value
		} else {
			// Collect field expression
			expression := p.parseExpressionRoot()
			expressionType := GetExpressionType(expression)

			// Check type
			if !orderedProperties[propertyIndex].dataType.CanBeAssigned(expressionType) {
				p.newError(expression.Position, "Property "+orderedPropertyNames[propertyIndex]+" of struct "+structName.Value+" has type "+orderedProperties[propertyIndex].dataType.String()+", but was assigned expression of type "+expressionType.String()+".")
			}

			// Store property
			propertyValues[propertyIndex] = expression
			propertyIndex++
		}

		// Consume comma
		if p.peek().TokenType == lexer.TT_DL_Comma {
			p.consume()
		}

		// Consume EOCs
		p.consumeEOCs()
	}

	if propertyIndex < len(properties) {
		p.newError(structName.Position, "Struct "+structName.Value+fmt.Sprintf(" has %d fields, but only %d fields were assigned.", len(properties), propertyIndex))
	}

	return propertyValues
}

func combineLiteralNodes(left, right *Node, parentNodeType NodeType) *Node {
	leftLiteral := left.Value.(*LiteralNode)
	rightLiteral := right.Value.(*LiteralNode)

	switch leftLiteral.PrimitiveType {
	// Booleans
	case data.DT_Bool:
		switch parentNodeType {
		case NT_Equal:
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) == rightLiteral.Value.(bool)}}
		case NT_NotEqual:
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) != rightLiteral.Value.(bool)}}
		case NT_And:
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) && rightLiteral.Value.(bool)}}
		case NT_Or:
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, leftLiteral.Value.(bool) || rightLiteral.Value.(bool)}}
		}
	// Integers
	case data.DT_Int:
		var value LiteralValue = nil

		// Arithmetic operations
		switch parentNodeType {
		case NT_Add:
			value = leftLiteral.Value.(int64) + rightLiteral.Value.(int64)
		case NT_Subtract:
			value = leftLiteral.Value.(int64) - rightLiteral.Value.(int64)
		case NT_Multiply:
			value = leftLiteral.Value.(int64) * rightLiteral.Value.(int64)
		case NT_Divide:
			value = leftLiteral.Value.(int64) / rightLiteral.Value.(int64)
		case NT_Power:
			value = VM.PowerInt64(leftLiteral.Value.(int64), rightLiteral.Value.(int64))
		case NT_Modulo:
			value = leftLiteral.Value.(int64) % rightLiteral.Value.(int64)
		}

		if value != nil {
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Int, value}}
		}

		// Comparison operators
		switch parentNodeType {
		case NT_Equal:
			value = leftLiteral.Value.(int64) == rightLiteral.Value.(int64)
		case NT_NotEqual:
			value = leftLiteral.Value.(int64) != rightLiteral.Value.(int64)
		case NT_Lower:
			value = leftLiteral.Value.(int64) < rightLiteral.Value.(int64)
		case NT_Greater:
			value = leftLiteral.Value.(int64) > rightLiteral.Value.(int64)
		case NT_LowerEqual:
			value = leftLiteral.Value.(int64) <= rightLiteral.Value.(int64)
		case NT_GreaterEqual:
			value = leftLiteral.Value.(int64) >= rightLiteral.Value.(int64)
		}

		if value != nil {
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, value}}
		}

	// Floats
	case data.DT_Float:
		var value LiteralValue = nil

		// Arithmetic operations
		switch parentNodeType {
		case NT_Add:
			value = leftLiteral.Value.(float64) + rightLiteral.Value.(float64)
		case NT_Subtract:
			value = leftLiteral.Value.(float64) - rightLiteral.Value.(float64)
		case NT_Multiply:
			value = leftLiteral.Value.(float64) * rightLiteral.Value.(float64)
		case NT_Divide:
			value = leftLiteral.Value.(float64) / rightLiteral.Value.(float64)
		case NT_Power:
			value = math.Pow(leftLiteral.Value.(float64), rightLiteral.Value.(float64))
		case NT_Modulo:
			value = math.Mod(leftLiteral.Value.(float64), rightLiteral.Value.(float64))
		}

		if value != nil {
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Float, value}}
		}

		// Comparison operators
		switch parentNodeType {
		case NT_Equal:
			value = leftLiteral.Value.(float64) == rightLiteral.Value.(float64)
		case NT_NotEqual:
			value = leftLiteral.Value.(float64) != rightLiteral.Value.(float64)
		case NT_Lower:
			value = leftLiteral.Value.(float64) < rightLiteral.Value.(float64)
		case NT_Greater:
			value = leftLiteral.Value.(float64) > rightLiteral.Value.(float64)
		case NT_LowerEqual:
			value = leftLiteral.Value.(float64) <= rightLiteral.Value.(float64)
		case NT_GreaterEqual:
			value = leftLiteral.Value.(float64) >= rightLiteral.Value.(float64)
		}

		if value != nil {
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_Bool, value}}
		}

	// Strings
	case data.DT_String:
		if parentNodeType == NT_Add {
			return &Node{left.Position.Combine(right.Position), NT_Literal, &LiteralNode{data.DT_String, left.Value.(*LiteralNode).Value.(string) + right.Value.(*LiteralNode).Value.(string)}}
		}
	}

	// Invalid operation, can't combine
	return &Node{left.Position.Combine(right.Position), parentNodeType, &TypedBinaryNode{left, right, &data.DataType{data.DT_Unknown, nil}}}
}
