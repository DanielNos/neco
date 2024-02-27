package codeGenerator

import (
	"fmt"
	data "neco/dataStructures"
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateExpression(node *parser.Node, loadLeft bool) {
	switch node.NodeType {
	// Literal
	case parser.NT_Literal:
		cg.generateLiteral(node, loadLeft)

	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)
		if loadLeft {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyReturnToOpA, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyReturnToOpB, NO_ARGS})
		}

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
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_NotOpA, NO_ARGS})
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
		cg.generateVariable(node.Value.(*parser.VariableNode).Identifier, loadLeft)

	// Lists
	case parser.NT_List:
		// Create list in ListA
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CreateListInListA, NO_ARGS})

		for _, node := range node.Value.(*parser.ListNode).Nodes {
			cg.generateExpression(node, true)
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_AppendOpAToListA, NO_ARGS})
		}

		if loadLeft {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyListAToOpA, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyListAToOpB, NO_ARGS})
		}

	// List values
	case parser.NT_ListValue:
		// OpB to OpStore
		if loadLeft {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpBToOpStore, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpAToOpStore, NO_ARGS})
		}

		// Generate index expression to OpA/OpB
		cg.generateExpression(node.Value.(*parser.ListValueNode).Index, loadLeft)

		// Create variable load instruction
		cg.generateVariable(node.Value.(*parser.ListValueNode).Identifier, loadLeft)

		// Change it's type to LoadListAtOpAToOpA/LoadListAtOpBToOpB
		if loadLeft {
			cg.instructions[len(cg.instructions)-1].InstructionType = VM.IT_LoadListAtOpAToOpA
		} else {
			cg.instructions[len(cg.instructions)-1].InstructionType = VM.IT_LoadListOpBToOpB
		}

		// Return data from OpStore to OpA/OpB
		if loadLeft {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpStoreToOpB, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpStoreToOpA, NO_ARGS})
		}

	case parser.NT_Not:
		cg.generateExpression(node.Value.(*parser.BinaryNode).Right, loadLeft)
		if loadLeft {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_NotOpA, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_NotOpB, NO_ARGS})
		}

	default:
		panic(fmt.Sprintf("Invalid node in generator expression: %s", node.NodeType))
	}
}

func (cg *CodeGenerator) generateExpressionArguments(binaryNode *parser.BinaryNode) {
	// Opearator on two leaf nodes
	if binaryNode.Left.NodeType.IsLeaf() && binaryNode.Right.NodeType.IsLeaf() {
		cg.generateExpression(binaryNode.Left, true)
		cg.generateExpression(binaryNode.Right, false)
		// Operator on left and leaf on right
	} else if binaryNode.Right.NodeType.IsLeaf() {
		cg.generateExpression(binaryNode.Left, true)
		cg.generateExpression(binaryNode.Right, false)
		// Operator on right and anything on left
	} else {
		cg.generateExpression(binaryNode.Right, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpAToOpStore, NO_ARGS})
		cg.generateExpression(binaryNode.Left, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpStoreToOpB, NO_ARGS})
	}
}

func (cg *CodeGenerator) generateVariable(variableName string, loadLeft bool) {
	identifier := cg.findVariableIdentifier(variableName)

	// Load variable to correct register
	if loadLeft {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LoadRegA, []byte{identifier}})
	} else {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LoadRegB, []byte{identifier}})
	}
}

func (cg *CodeGenerator) generateLiteral(node *parser.Node, loadLeft bool) {
	literalNode := node.Value.(*parser.LiteralNode)

	instruction := VM.IT_LoadConstRegA
	if !loadLeft {
		instruction = VM.IT_LoadConstRegB
	}

	switch literalNode.DType {
	// Bool
	case data.DT_Bool:
		if loadLeft {
			if literalNode.Value.(bool) {
				cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_SetRegATrue, NO_ARGS})
			} else {
				cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_SetRegAFalse, NO_ARGS})
			}
		} else {
			if literalNode.Value.(bool) {
				cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_SetRegBTrue, NO_ARGS})
			} else {
				cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_SetRegBFalse, NO_ARGS})
			}
		}

	// Int
	case data.DT_Int:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.intConstants[literalNode.Value.(int64)])}})

	// Float
	case data.DT_Float:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.floatConstants[literalNode.Value.(float64)])}})
	// String
	case data.DT_String:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.stringConstants[literalNode.Value.(string)])}})
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
		return expression.Value.(*parser.ListValueNode).ListSymbol.VariableType.SubType.(data.DataType)
	}

	panic(fmt.Sprintf("Can't determine expression data type from %s.", parser.NodeTypeToString[expression.NodeType]))
}
