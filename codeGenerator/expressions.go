package codeGenerator

import (
	data "neco/dataStructures"
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateExpression(node *parser.Node) {
	switch node.NodeType {
	// Literal
	case parser.NT_Literal:
		cg.generateLiteral(node)

	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	// Operators
	case parser.NT_Add, parser.NT_Subtract, parser.NT_Multiply, parser.NT_Divide, parser.NT_Power, parser.NT_Modulo:
		// Generate left side
		binaryNode := node.Value.(*parser.TypedBinaryNode)
		cg.generateExpression(binaryNode.Left)

		// Insert elements to a set (elements are inserted by themselves, we don't create another set)
		if binaryNode.DataType.Type == data.DT_Set {
			for _, element := range binaryNode.Right.Value.(*parser.ListNode).Nodes {
				cg.generateExpression(element)
				cg.addInstruction(VM.IT_InsertToSet)
			}
			break
		}

		cg.generateExpression(binaryNode.Right)

		// Generate operator
		// Concatenate strings
		if binaryNode.DataType.Type == data.DT_String {
			cg.addInstruction(VM.IT_StringConcat)
			// Concatenate lists
		} else if binaryNode.DataType.Type == data.DT_List {
			cg.addInstruction(VM.IT_ListConcat)
			// Operation on ints
		} else if binaryNode.DataType.Type == data.DT_Int {
			cg.addInstruction(intOperatorToInstruction[node.NodeType])
			// Operation on floats
		} else {
			cg.addInstruction(floatOperatorToInstruction[node.NodeType])
		}

	// Logical operators
	case parser.NT_And:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		cg.addInstruction(VM.IT_And)

	case parser.NT_Or:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		cg.addInstruction(VM.IT_Or)

	// Comparison operators
	case parser.NT_Equal, parser.NT_NotEqual:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)

		// Generate operator
		cg.addInstruction(VM.IT_Equal)

		if node.NodeType == parser.NT_NotEqual {
			cg.addInstruction(VM.IT_Not)
		}

	case parser.NT_Lower, parser.NT_Greater, parser.NT_LowerEqual, parser.NT_GreaterEqual:
		// Generate arguments
		binaryNode := node.Value.(*parser.TypedBinaryNode)
		cg.generateExpression(binaryNode.Left)
		cg.generateExpression(binaryNode.Right)

		// Generate operator
		leftType := getExpressionType(binaryNode.Left)

		// Compare ints
		if leftType.Type == data.DT_Int {
			cg.addInstruction(comparisonOperatorToIntInstruction[node.NodeType])
			// Compare floats
		} else if leftType.Type == data.DT_Float {
			cg.addInstruction(comparisonOperatorToFloatInstruction[node.NodeType])
		} else {
			panic("Can't generate comparision instruction on operator nodes.")
		}

	// Variables
	case parser.NT_Variable:
		cg.generateVariable(node.Value.(*parser.VariableNode).Identifier)

	// Lists
	case parser.NT_List:
		// Create list
		cg.addInstruction(VM.IT_CreateList)

		// Append elements
		for _, node := range node.Value.(*parser.ListNode).Nodes {
			cg.generateExpression(node)
			cg.addInstruction(VM.IT_AppendToList)
		}

	// List values
	case parser.NT_ListValue:
		// Generate list expression
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)

		// Generate index expression
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)

		// Generate indexing instruction
		if parser.GetExpressionType(node.Value.(*parser.TypedBinaryNode).Left).Type == data.DT_String {
			cg.addInstruction(VM.IT_IndexString)
		} else {
			cg.addInstruction(VM.IT_IndexList)
		}

	// Logical not
	case parser.NT_Not:
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		cg.addInstruction(VM.IT_Not)

	// Enums
	case parser.NT_Enum:
		cg.addInstruction(VM.IT_LoadConst, uint8(cg.intConstants[node.Value.(*parser.EnumNode).Value]))

	// Struct objects
	case parser.NT_Object:
		ObjectNode := node.Value.(*parser.ObjectNode)

		// Create object
		cg.addInstruction(VM.IT_CreateObject, byte(cg.stringConstants[ObjectNode.Identifier]))

		// Generate properties
		for _, property := range ObjectNode.Properties {
			cg.generateExpression(property)
			cg.addInstruction(VM.IT_AddField)
		}

	// Struct fields
	case parser.NT_ObjectField:
		objectFieldNode := node.Value.(*parser.ObjectFieldNode)

		cg.generateExpression(objectFieldNode.Object)

		cg.addInstruction(VM.IT_GetField, byte(objectFieldNode.FieldIndex))

	// Set literals
	case parser.NT_Set:
		cg.addInstruction(VM.IT_CreateSet)

		elements := node.Value.(*parser.ListNode).Nodes
		usedElements := map[any]struct{}{}

		for _, element := range elements {
			// Skip literals that were already inserted
			if cg.optimize && element.NodeType == parser.NT_Literal {
				_, exists := usedElements[element.Value.(*parser.LiteralNode).Value]
				if exists {
					continue
				}
				usedElements[element.Value.(*parser.LiteralNode).Value] = struct{}{}
			}

			// Genearate expression and insertion
			cg.generateExpression(element)
			cg.addInstruction(VM.IT_InsertToSet)
		}

	// Set contains
	case parser.NT_In:
		binaryNode := node.Value.(*parser.TypedBinaryNode)
		cg.generateExpression(binaryNode.Right)
		cg.generateExpression(binaryNode.Left)
		cg.addInstruction(VM.IT_SetContains)

	default:
		panic("Invalid node in generator expression: " + node.NodeType.String())
	}
}

func (cg *CodeGenerator) generateVariable(variableName string) {
	identifier := cg.findVariableIdentifier(variableName)
	cg.addInstruction(VM.IT_Load, identifier)
}

func (cg *CodeGenerator) generateLiteral(node *parser.Node) {
	literalNode := node.Value.(*parser.LiteralNode)

	switch literalNode.PrimitiveType {
	// Bool
	case data.DT_Bool:
		if literalNode.Value.(bool) {
			cg.addInstruction(VM.IT_PushTrue)
		} else {
			cg.addInstruction(VM.IT_PushFalse)
		}

	// Int
	case data.DT_Int:
		cg.addInstruction(VM.IT_LoadConst, uint8(cg.intConstants[literalNode.Value.(int64)]))

	// Float
	case data.DT_Float:
		cg.addInstruction(VM.IT_LoadConst, uint8(cg.floatConstants[literalNode.Value.(float64)]))

	// String
	case data.DT_String:
		cg.addInstruction(VM.IT_LoadConst, uint8(cg.stringConstants[literalNode.Value.(string)]))

	// None
	case data.DT_None:
		cg.addInstruction(VM.IT_PushNone)
	}
}

func getExpressionType(expression *parser.Node) *data.DataType {
	if expression.NodeType.IsOperator() {
		// Unary operator
		if expression.Value.(*parser.TypedBinaryNode).Left == nil {
			unaryType := getExpressionType(expression.Value.(*parser.TypedBinaryNode).Right)
			return unaryType
		}

		// Binary operator
		leftType := getExpressionType(expression.Value.(*parser.TypedBinaryNode).Left)
		rightType := getExpressionType(expression.Value.(*parser.TypedBinaryNode).Right)

		return &data.DataType{max(leftType.Type, rightType.Type), nil}
	}

	switch expression.NodeType {
	case parser.NT_Literal:
		return &data.DataType{expression.Value.(*parser.LiteralNode).PrimitiveType, nil}
	case parser.NT_Variable:
		return expression.Value.(*parser.VariableNode).DataType
	case parser.NT_FunctionCall:
		return expression.Value.(*parser.FunctionCallNode).ReturnType
	case parser.NT_List:
		return expression.Value.(*parser.ListNode).DataType
	case parser.NT_ListValue:
		return getExpressionType(expression.Value.(*parser.TypedBinaryNode).Left).SubType.(*data.DataType)
	}

	panic("Can't determine expression data type from " + parser.NodeTypeToString[expression.NodeType] + ".")
}
