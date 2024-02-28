package codeGenerator

import (
	"fmt"
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
		// Generate arguments
		binaryNode := node.Value.(*parser.BinaryNode)
		cg.generateExpressionArguments(binaryNode)

		// Generate operator
		// Concatenate strings
		if binaryNode.DataType.DType == data.DT_String {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StringConcat, NO_ARGS})
			// Concatenate lists
		} else if binaryNode.DataType.DType == data.DT_List {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_ListConcat, NO_ARGS})
			// Operation on ints
		} else if binaryNode.DataType.DType == data.DT_Int {
			cg.instructions = append(cg.instructions, VM.Instruction{intOperatorToInstruction[node.NodeType], NO_ARGS})
			// Operation on floats
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{floatOperatorToInstruction[node.NodeType], NO_ARGS})
		}

	// Comparison operators
	case parser.NT_Equal, parser.NT_NotEqual:
		// Generate arguments
		binaryNode := node.Value.(*parser.BinaryNode)
		cg.generateExpressionArguments(binaryNode)

		// Generate operator
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Equal, NO_ARGS})

		if node.NodeType == parser.NT_NotEqual {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Not, NO_ARGS})
		}

	case parser.NT_Lower, parser.NT_Greater, parser.NT_LowerEqual, parser.NT_GreaterEqual:
		// Generate arguments
		binaryNode := node.Value.(*parser.BinaryNode)
		cg.generateExpressionArguments(binaryNode)

		// Generate operator
		leftType := getExpressionType(binaryNode.Left)

		// Compare ints
		if leftType.DType == data.DT_Int {
			cg.instructions = append(cg.instructions, VM.Instruction{comparisonOperatorToIntInstruction[node.NodeType], NO_ARGS})
			// Compare floats
		} else if leftType.DType == data.DT_Float {
			cg.instructions = append(cg.instructions, VM.Instruction{comparisonOperatorToFloatInstruction[node.NodeType], NO_ARGS})
		} else {
			panic("Can't generate comparision instruction on operator nodes.")
		}

	// Variables
	case parser.NT_Variable:
		cg.generateVariable(node.Value.(*parser.VariableNode).Identifier)

	// Lists
	case parser.NT_List:
		// Create list in ListA
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CreateList, NO_ARGS})

		for _, node := range node.Value.(*parser.ListNode).Nodes {
			cg.generateExpression(node)
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_AppendToList, NO_ARGS})
		}

	// List values
	case parser.NT_ListValue:
		// Generate list expression
		cg.generateExpression(node.Value.(*parser.BinaryNode).Left)

		// Generate index expression
		cg.generateExpression(node.Value.(*parser.BinaryNode).Right)

		// Generate LoadListAt instruction
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_IndexList, NO_ARGS})

	case parser.NT_Not:
		cg.generateExpression(node.Value.(*parser.BinaryNode).Right)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Not, NO_ARGS})

	default:
		panic(fmt.Sprintf("Invalid node in generator expression: %s", node.NodeType))
	}
}

func (cg *CodeGenerator) generateExpressionArguments(binaryNode *parser.BinaryNode) {
	cg.generateExpression(binaryNode.Left)
	cg.generateExpression(binaryNode.Right)
}

func (cg *CodeGenerator) generateVariable(variableName string) {
	identifier := cg.findVariableIdentifier(variableName)
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Load, []byte{identifier}})
}

func (cg *CodeGenerator) generateLiteral(node *parser.Node) {
	literalNode := node.Value.(*parser.LiteralNode)

	switch literalNode.DType {
	// Bool
	case data.DT_Bool:
		if literalNode.Value.(bool) {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PushTrue, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PushFalse, NO_ARGS})
		}

	// Int
	case data.DT_Int:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.intConstants[literalNode.Value.(int64)])}})

	// Float
	case data.DT_Float:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.floatConstants[literalNode.Value.(float64)])}})

	// String
	case data.DT_String:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LoadConst, []byte{uint8(cg.stringConstants[literalNode.Value.(string)])}})
	}
}

func getExpressionType(expression *parser.Node) data.DataType {
	if expression.NodeType.IsOperator() {
		// Unary operator
		if expression.Value.(*parser.BinaryNode).Left == nil {
			unaryType := getExpressionType(expression.Value.(*parser.BinaryNode).Right)
			return unaryType
		}

		// Binary operator
		leftType := getExpressionType(expression.Value.(*parser.BinaryNode).Left)
		rightType := getExpressionType(expression.Value.(*parser.BinaryNode).Right)

		return data.DataType{max(leftType.DType, rightType.DType), nil}
	}

	switch expression.NodeType {
	case parser.NT_Literal:
		return data.DataType{expression.Value.(*parser.LiteralNode).DType, nil}
	case parser.NT_Variable:
		return expression.Value.(*parser.VariableNode).DataType
	case parser.NT_FunctionCall:
		return *expression.Value.(*parser.FunctionCallNode).ReturnType
	case parser.NT_List:
		return expression.Value.(*parser.ListNode).DataType
	case parser.NT_ListValue:
		return getExpressionType(expression.Value.(*parser.BinaryNode).Left)
	}

	panic(fmt.Sprintf("Can't determine expression data type from %s.", parser.NodeTypeToString[expression.NodeType]))
}
