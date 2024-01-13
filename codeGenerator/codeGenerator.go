package codeGenerator

import (
	"fmt"
	"math"
	"neko/errors"
	"neko/logger"
	"neko/parser"
	VM "neko/virtualMachine"
)

const EMPTY byte = 0

type CodeGenerator struct {
	filePath string
	tree     *parser.Node

	Constants       []*parser.LiteralNode
	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int

	instructions []VM.Instruction

	line uint

	ErrorCount int
}

func NewGenerator(tree *parser.Node, outputFile string) *CodeGenerator {
	return &CodeGenerator{outputFile, tree, []*parser.LiteralNode{}, map[int64]int{}, map[float64]int{}, map[string]int{}, []VM.Instruction{}, 0, 0}
}

func (cg *CodeGenerator) Generate() *[]VM.Instruction {
	statements := cg.tree.Value.(*parser.ModuleNode).Statements.Statements

	cg.line = statements[0].Position.Line
	cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: cg.toByte(cg.line), InstructionValue: []byte{}})

	for _, node := range statements {
		cg.generateNode(node)
	}

	return &cg.instructions
}

func (cg *CodeGenerator) newError(message string) {
	logger.Error(message)
	cg.ErrorCount++

	if cg.ErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.ERROR_CODE_GENERATION, fmt.Sprintf("Failed code generation with %d errors.", cg.ErrorCount))
	}
}

func (cg *CodeGenerator) generateNode(node *parser.Node) {
	if node.Position.Line > cg.line {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: cg.toByte(node.Position.Line - cg.line), InstructionValue: []byte{}})
		cg.line = node.Position.Line
	}

	switch node.NodeType {
	case parser.NT_FunctionDeclare:
		if node.Value.(*parser.FunctionDeclareNode).Identifier == "entry" {
			cg.generateBody(node.Value.(*parser.FunctionDeclareNode))
		}
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)
	}
}

func (cg *CodeGenerator) generateBody(functionNode *parser.FunctionDeclareNode) {
	for _, node := range functionNode.Body.Value.(*parser.ScopeNode).Statements {
		cg.generateNode(node)
	}

	cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{0}})
}

func (cg *CodeGenerator) generateFunctionCall(node *parser.Node) {
	functionCall := node.Value.(*parser.FunctionCallNode)
	cg.generateArguments(functionCall.Arguments)

	builtInFunction, exists := builtInFunctions[functionCall.Identifier]

	if exists {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunction, InstructionValue: []byte{builtInFunction}})
	}
}

func (cg *CodeGenerator) generateArguments(arguments []*parser.Node) {
	for _, argument := range arguments {
		cg.generateExpression(argument)
		if cg.instructions[len(cg.instructions)-1].InstructionType != VM.IT_LoadConstant {
			cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Push, InstructionValue: []byte{VM.Reg_GenericA, VM.Stack_Argument}})
		}
	}
}

func (cg *CodeGenerator) generateExpression(node *parser.Node) {
	switch node.NodeType {
	case parser.NT_Literal:
		cg.generateLiteral(node)

	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	case parser.NT_Add, parser.NT_Subtract, parser.NT_Multiply, parser.NT_Divide, parser.NT_Power, parser.NT_Modulo:
		binaryNode := node.Value.(*parser.BinaryNode)

		cg.generateExpression(binaryNode.Left)
		cg.generateExpression(binaryNode.Right)

		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: nodeTypeToInstructionTypeInt[node.NodeType], InstructionValue: []byte{}})

	default:
		panic("Invalid node in generator expression!")
	}
}

func (cg *CodeGenerator) generateLiteral(node *parser.Node) {
	literalNode := node.Value.(*parser.LiteralNode)

	switch literalNode.DataType {
	case parser.DT_String:
		constantIndex, exists := cg.stringConstants[literalNode.Value.(string)]

		if !exists {
			cg.Constants = append(cg.Constants, literalNode)
			constantIndex = len(cg.Constants) - 1
			cg.stringConstants[literalNode.Value.(string)] = constantIndex

			if constantIndex == math.MaxUint8 {
				logger.Error(fmt.Sprintf("Constant pool overflow. There are more than %d constants.", math.MaxUint8))
			}
		}

		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_LoadConstant, InstructionValue: []byte{uint8(constantIndex), VM.Stack_Argument}})

	case parser.DT_Int:

		constantIndex, exists := cg.intConstants[literalNode.Value.(int64)]

		if !exists {
			cg.Constants = append(cg.Constants, literalNode)
			constantIndex = len(cg.Constants) - 1
			cg.intConstants[literalNode.Value.(int64)] = constantIndex

			if constantIndex == math.MaxUint8 {
				logger.Error(fmt.Sprintf("Constant pool overflow. There are more than %d constants.", math.MaxUint8))
			}
		}

		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_LoadConstant, InstructionValue: []byte{uint8(constantIndex), VM.Stack_Argument}})

	case parser.DT_Float:

		constantIndex, exists := cg.floatConstants[literalNode.Value.(float64)]

		if !exists {
			cg.Constants = append(cg.Constants, literalNode)
			constantIndex = len(cg.Constants) - 1
			cg.floatConstants[literalNode.Value.(float64)] = constantIndex

			if constantIndex == math.MaxUint8 {
				logger.Error(fmt.Sprintf("Constant pool overflow. There are more than %d constants.", math.MaxUint8))
			}
		}

		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_LoadConstant, InstructionValue: []byte{uint8(constantIndex), VM.Stack_Argument}})

	}
}
