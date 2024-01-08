package codegenerator

import (
	"fmt"
	"math"
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
}

func NewGenerator(tree *parser.Node, outputFile string) *CodeGenerator {
	return &CodeGenerator{outputFile, tree, []*parser.LiteralNode{}, map[int64]int{}, map[float64]int{}, map[string]int{}, []VM.Instruction{}}
}

func (cg *CodeGenerator) Generate() *[]VM.Instruction {
	for _, node := range cg.tree.Value.(*parser.ModuleNode).Statements.Statements {
		cg.generateNode(node)
	}

	return &cg.instructions
}

func (cg *CodeGenerator) generateNode(node *parser.Node) {
	switch node.NodeType {
	case parser.NT_FunctionDeclare:
		if node.Value.(*parser.FunctionDeclareNode).Identifier == "entry" {
			cg.generateBody(node.Value.(*parser.FunctionDeclareNode))
		}
	case parser.NT_FunctionCall:
		functionCall := node.Value.(*parser.FunctionCallNode)
		cg.generateArguments(functionCall.Arguments)

		if functionCall.Identifier == "print" {
			cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunction, InstructionValue: []byte{VM.BIF_Print}})
		} else if functionCall.Identifier == "printLine" {
			cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunction, InstructionValue: []byte{VM.BIF_PrintLine}})
		}
	}
}

func (cg *CodeGenerator) generateBody(functionNode *parser.FunctionDeclareNode) {
	for _, node := range functionNode.Body.Value.(*parser.ScopeNode).Statements {
		cg.generateNode(node)
	}

	cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{0}})
}

func (cg *CodeGenerator) generateArguments(arguments []*parser.Node) {
	for _, argument := range arguments {
		cg.generateExpression(argument)
	}
}

func (cg *CodeGenerator) generateExpression(node *parser.Node) {
	switch node.NodeType {
	case parser.NT_Literal:
		cg.generateLiteral(node)
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
		stringIndex, exists := cg.stringConstants[literalNode.Value.(string)]

		if !exists {
			cg.Constants = append(cg.Constants, literalNode)
			stringIndex = len(cg.Constants) - 1
			cg.stringConstants[literalNode.Value.(string)] = stringIndex

			if stringIndex == math.MaxUint16 {
				logger.Error(fmt.Sprintf("Constant pool overflow. There are more than %d constants.", math.MaxUint16))
			}
		}

		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_LoadConstant, InstructionValue: []byte{uint8(stringIndex), VM.Stack_Argument}})
	}
}
