package codeGenerator

import (
	"fmt"
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
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyRegDToA, NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyRegDToB, NO_ARGS})
		}

	// Operators
	case parser.NT_Add, parser.NT_Subtract, parser.NT_Multiply, parser.NT_Divide, parser.NT_Power, parser.NT_Modulo:
		// Generate arguments
		binaryNode := node.Value.(*parser.BinaryNode)
		cg.generateExpressionArguments(binaryNode)

		// Generate operator
		// Concatenate strings
		if binaryNode.DataType == parser.DT_String {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StringConcat, NO_ARGS})
			// Operation on ints
		} else if binaryNode.DataType == parser.DT_Int {
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
		if binaryNode.DataType == parser.DT_Int {
			cg.instructions = append(cg.instructions, VM.Instruction{logicalOperatorToIntInstruction[node.NodeType], NO_ARGS})
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{logicalOperatorToFloatInstruction[node.NodeType], NO_ARGS})
		}

	// Variables
	case parser.NT_Variable:
		cg.generateVariable(node, loadLeft)

	default:
		panic(fmt.Sprintf("Invalid node in generator expression: %s", node.NodeType))
	}
}

func (cg *CodeGenerator) generateExpressionArguments(binaryNode *parser.BinaryNode) {
	// Opearator on two leaf nodes
	if (binaryNode.Left.NodeType == parser.NT_Variable || binaryNode.Left.NodeType == parser.NT_Literal) && (binaryNode.Right.NodeType == parser.NT_Variable || binaryNode.Right.NodeType == parser.NT_Literal) {
		cg.generateExpression(binaryNode.Left, true)
		cg.generateExpression(binaryNode.Right, false)
		// Operator on left and leaf on right
	} else if binaryNode.Right.NodeType == parser.NT_Variable || binaryNode.Right.NodeType == parser.NT_Literal {
		cg.generateExpression(binaryNode.Left, true)
		cg.generateExpression(binaryNode.Right, false)
		// Operator on right and anything on left
	} else {
		cg.generateExpression(binaryNode.Right, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyRegAToC, NO_ARGS})
		cg.generateExpression(binaryNode.Left, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyRegCToB, NO_ARGS})
	}
}

func (cg *CodeGenerator) generateVariable(node *parser.Node, loadLeft bool) {
	identifier := cg.variableIdentifiers.Top.Value.(map[string]uint8)[node.Value.(*parser.VariableNode).Identifier]

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

	switch literalNode.DataType {
	// Bool
	case parser.DT_Bool:
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
	case parser.DT_Int:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.intConstants[literalNode.Value.(int64)])}})

	// Float
	case parser.DT_Float:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.floatConstants[literalNode.Value.(float64)])}})
	// String
	case parser.DT_String:
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(cg.stringConstants[literalNode.Value.(string)])}})
	}
}
