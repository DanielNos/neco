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
				*cg.target = append(*cg.target, VM.Instruction{VM.IT_InsertToSet, NO_ARGS})
			}
			break
		}

		cg.generateExpression(binaryNode.Right)

		// Generate operator
		// Concatenate strings
		if binaryNode.DataType.Type == data.DT_String {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_StringConcat, NO_ARGS})
			// Concatenate lists
		} else if binaryNode.DataType.Type == data.DT_List {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_ListConcat, NO_ARGS})
			// Operation on ints
		} else if binaryNode.DataType.Type == data.DT_Int {
			*cg.target = append(*cg.target, VM.Instruction{intOperatorToInstruction[node.NodeType], NO_ARGS})
			// Operation on floats
		} else {
			*cg.target = append(*cg.target, VM.Instruction{floatOperatorToInstruction[node.NodeType], NO_ARGS})
		}

	// Logical operators
	case parser.NT_And:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_And, NO_ARGS})

	case parser.NT_Or:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_Or, NO_ARGS})

	// Comparison operators
	case parser.NT_Equal, parser.NT_NotEqual:
		// Generate arguments
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)

		// Generate operator
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_Equal, NO_ARGS})

		if node.NodeType == parser.NT_NotEqual {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_Not, NO_ARGS})
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
			*cg.target = append(*cg.target, VM.Instruction{comparisonOperatorToIntInstruction[node.NodeType], NO_ARGS})
			// Compare floats
		} else if leftType.Type == data.DT_Float {
			*cg.target = append(*cg.target, VM.Instruction{comparisonOperatorToFloatInstruction[node.NodeType], NO_ARGS})
		} else {
			panic("Can't generate comparision instruction on operator nodes.")
		}

	// Variables
	case parser.NT_Variable:
		cg.generateVariable(node.Value.(*parser.VariableNode).Identifier)

	// Lists
	case parser.NT_List:
		// Create list
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_CreateList, NO_ARGS})

		// Append elements
		for _, node := range node.Value.(*parser.ListNode).Nodes {
			cg.generateExpression(node)
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_AppendToList, NO_ARGS})
		}

	// List values
	case parser.NT_ListValue:
		// Generate list expression
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Left)

		// Generate index expression
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)

		// Generate indexing instruction
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_IndexList, NO_ARGS})

	// Logical not
	case parser.NT_Not:
		cg.generateExpression(node.Value.(*parser.TypedBinaryNode).Right)
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_Not, NO_ARGS})

	// Enums
	case parser.NT_Enum:
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.intConstants[node.Value.(*parser.EnumNode).Value])}})

	// Struct objects
	case parser.NT_Object:
		ObjectNode := node.Value.(*parser.ObjectNode)

		// Create object
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_CreateObject, []byte{byte(cg.stringConstants[ObjectNode.Identifier])}})

		// Generate properties
		for _, property := range ObjectNode.Properties {
			cg.generateExpression(property)
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_StoreField, NO_ARGS})
		}

	// Struct fields
	case parser.NT_StructField:
		ObjectFieldNode := node.Value.(*parser.ObjectFieldNode)

		cg.generateVariable(ObjectFieldNode.Identifier)

		*cg.target = append(*cg.target, VM.Instruction{VM.IT_GetField, []byte{byte(ObjectFieldNode.PropertyIndex)}})

	// Set literals
	case parser.NT_Set:
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_CreateSet, NO_ARGS})

		elements := node.Value.(*parser.ListNode).Nodes
		usedElements := map[interface{}]struct{}{}

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
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_InsertToSet, NO_ARGS})
		}

	// Set contains
	case parser.NT_In:
		binaryNode := node.Value.(*parser.TypedBinaryNode)
		cg.generateExpression(binaryNode.Right)
		cg.generateExpression(binaryNode.Left)
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_SetContains, NO_ARGS})

	default:
		panic("Invalid node in generator expression: " + node.NodeType.String())
	}
}

func (cg *CodeGenerator) generateVariable(variableName string) {
	identifier := cg.findVariableIdentifier(variableName)
	*cg.target = append(*cg.target, VM.Instruction{VM.IT_Load, []byte{identifier}})
}

func (cg *CodeGenerator) generateLiteral(node *parser.Node) {
	literalNode := node.Value.(*parser.LiteralNode)

	switch literalNode.PrimitiveType {
	// Bool
	case data.DT_Bool:
		if literalNode.Value.(bool) {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_PushTrue, NO_ARGS})
		} else {
			*cg.target = append(*cg.target, VM.Instruction{VM.IT_PushFalse, NO_ARGS})
		}

	// Int
	case data.DT_Int:
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.intConstants[literalNode.Value.(int64)])}})

	// Float
	case data.DT_Float:
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.floatConstants[literalNode.Value.(float64)])}})

	// String
	case data.DT_String:
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.stringConstants[literalNode.Value.(string)])}})
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
