package codeGenerator

import (
	"fmt"
	"math"
	"neco/dataStructures"
	"neco/errors"
	"neco/logger"
	"neco/parser"
	VM "neco/virtualMachine"
)

var NO_ARGS = []byte{}

type CodeGenerator struct {
	filePath string
	tree     *parser.Node

	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int
	constants       []interface{}

	instructions []VM.Instruction

	variableIdentifierCounters *dataStructures.Stack
	variableIdentifiers        *dataStructures.Stack

	line uint

	ErrorCount int
}

func NewGenerator(tree *parser.Node, outputFile string, intConstants map[int64]int, floatConstants map[float64]int, stringConstants map[string]int) *CodeGenerator {
	codeGenerator := &CodeGenerator{outputFile, tree, intConstants, floatConstants, stringConstants, make([]interface{}, len(intConstants)+len(floatConstants)+len(stringConstants)), []VM.Instruction{}, dataStructures.NewStack(), dataStructures.NewStack(), 1, 0}

	codeGenerator.variableIdentifierCounters.Push(uint8(0))
	codeGenerator.variableIdentifiers.Push(map[string]uint8{})

	return codeGenerator
}

func (cg *CodeGenerator) Generate() *[]VM.Instruction {
	// Generate constant IDs
	cg.generateConstantIDs()

	// Get first line
	statements := cg.tree.Value.(*parser.ModuleNode).Statements.Statements
	cg.line = statements[0].Position.Line

	// Generate instructions
	for _, node := range statements {
		cg.generateNode(node)
	}

	return &cg.instructions
}

func (cg *CodeGenerator) newError(message string) {
	logger.Error(message)
	cg.ErrorCount++

	if cg.ErrorCount > errors.MAX_ERROR_COUNT {
		logger.Fatal(errors.CODE_GENERATION, fmt.Sprintf("Failed code generation with %d errors.", cg.ErrorCount))
	}
}

func (cg *CodeGenerator) generateConstantIDs() {
	// Map constant values to their index in global constant table.
	// This table is sorted by type, in order: strings, ints, floats.

	id := 0

	// Strings
	for key := range cg.stringConstants {
		cg.constants[id] = key
		cg.stringConstants[key] = id
		id++
	}

	// Integers
	for key := range cg.intConstants {
		cg.constants[id] = key
		cg.intConstants[key] = id
		id++
	}

	// Floats
	for key := range cg.floatConstants {
		cg.constants[id] = key
		cg.floatConstants[key] = id
		id++
	}

	// More than 255 constants
	if id >= math.MaxUint8 {
		logger.Error(fmt.Sprintf("Constant pool overflow with %d constants. Constant pool can only contain maximum of %d constants.", id, math.MaxUint8))
	}
}

func (cg *CodeGenerator) generateNode(node *parser.Node) {
	if node.Position.Line > cg.line {
		cg.instructions = append(cg.instructions, VM.Instruction{cg.lineToByte(node.Position.Line - cg.line - 1), NO_ARGS})
		cg.line = node.Position.Line
	}

	switch node.NodeType {
	// Function declaration
	case parser.NT_FunctionDeclare:
		if node.Value.(*parser.FunctionDeclareNode).Identifier == "entry" {
			cg.generateBody(node.Value.(*parser.FunctionDeclareNode))
		}

	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	// Variable declaration
	case parser.NT_VariableDeclare:
		cg.generateVariableDeclare(node)

	// Assignment
	case parser.NT_Assign:
		assignNode := node.Value.(*parser.AssignNode)

		cg.generateExpression(assignNode.Expression, true)

		identifier := cg.variableIdentifiers.Top.Value.(map[string]uint8)[assignNode.Identifier]
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StoreRegA, []byte{identifier}})
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
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunc, InstructionValue: []byte{builtInFunction}})
	}
}

func (cg *CodeGenerator) generateArguments(arguments []*parser.Node) {
	for _, argument := range arguments {
		cg.generateExpression(argument, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PushRegAToArgStack, NO_ARGS})
	}
}

func (cg *CodeGenerator) generateVariableDeclare(node *parser.Node) {
	variable := node.Value.(*parser.VariableDeclareNode)

	for i := 0; i < len(variable.Identifiers); i++ {
		cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]] = cg.variableIdentifierCounters.Top.Value.(uint8)
		cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1

		switch variable.VariableType.DataType {
		case parser.DT_Bool:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareBool, NO_ARGS})
		case parser.DT_Int:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareInt, NO_ARGS})
		case parser.DT_Float:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareFloat, NO_ARGS})
		case parser.DT_String:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareString, NO_ARGS})
		}
	}
}

func (cg *CodeGenerator) generateExpression(node *parser.Node, loadLeft bool) {
	switch node.NodeType {
	// Literal
	case parser.NT_Literal:
		cg.generateLiteral(node, loadLeft)

	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	// Operators
	case parser.NT_Add, parser.NT_Subtract, parser.NT_Multiply, parser.NT_Divide, parser.NT_Power, parser.NT_Modulo:
		binaryNode := node.Value.(*parser.BinaryNode)

		if !binaryNode.Left.NodeType.IsOperator() && !binaryNode.Right.NodeType.IsOperator() {
			cg.generateExpression(binaryNode.Left, true)
			cg.generateExpression(binaryNode.Right, false)
		}

		// Add strings
		if binaryNode.DataType == parser.DT_String {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StringConcat, NO_ARGS})
			// Add ints
		} else {
			cg.instructions = append(cg.instructions, VM.Instruction{nodeTypeToInstructionTypeInt[node.NodeType], NO_ARGS})
		}
	// Variable
	case parser.NT_Variable:
		cg.generateVariable(node, loadLeft)

	default:
		panic(fmt.Sprintf("Invalid node in generator expression: %s", node.NodeType))
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

	// String
	case parser.DT_String:
		constantIndex, _ := cg.stringConstants[literalNode.Value.(string)]
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(constantIndex)}})

	// Int
	case parser.DT_Int:
		constantIndex, _ := cg.intConstants[literalNode.Value.(int64)]
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(constantIndex)}})

	// Float
	case parser.DT_Float:
		constantIndex, _ := cg.floatConstants[literalNode.Value.(float64)]
		cg.instructions = append(cg.instructions, VM.Instruction{instruction, []byte{uint8(constantIndex)}})
	}
}
