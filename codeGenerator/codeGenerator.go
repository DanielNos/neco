package codeGenerator

import (
	"fmt"
	"math"
	"neco/codeOptimizer"
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
	optimize bool

	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int
	constants       []interface{}

	instructions []VM.Instruction

	variableIdentifierCounters *dataStructures.Stack
	variableIdentifiers        *dataStructures.Stack

	functions []int // Function number : function start

	line uint

	ErrorCount int
}

func NewGenerator(tree *parser.Node, outputFile string, intConstants map[int64]int, floatConstants map[float64]int, stringConstants map[string]int, optimize bool) *CodeGenerator {
	codeGenerator := &CodeGenerator{
		outputFile,
		tree,
		optimize,

		intConstants,
		floatConstants,
		stringConstants,
		make([]interface{}, len(intConstants)+len(floatConstants)+len(stringConstants)),

		[]VM.Instruction{},

		dataStructures.NewStack(),
		dataStructures.NewStack(),

		[]int{},
		0,
		0,
	}

	codeGenerator.variableIdentifierCounters.Push(uint8(0))
	codeGenerator.variableIdentifiers.Push(map[string]uint8{})

	return codeGenerator
}

func (cg *CodeGenerator) Generate() *[]VM.Instruction {
	// Generate constant IDs
	cg.generateConstantIDs()

	// Get root statement list
	statements := cg.tree.Value.(*parser.ModuleNode).Statements.Statements

	// No instructions, generate line offset and halt instruction
	if len(statements) == 0 {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{0}})
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Halt, []byte{0}})

		logger.Warning("Source code doesn't contain any symbols. Binary will be generated, but will contain no instructions.")

		return &cg.instructions
	}

	// Generate functions
	for _, node := range statements {
		if node.NodeType == parser.NT_FunctionDeclare {
			if cg.line < node.Position.Line {
				cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{byte(node.Position.Line - cg.line)}})
				cg.line = node.Position.Line
			}

			cg.generateFunction(node.Value.(*parser.FunctionDeclareNode))
		}
	}

	// Optimize instructions
	if !cg.optimize {
		codeOptimizer.Optimize(cg.instructions)
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
	// If node line has changed, insert line offset instruction
	if node.Position.Line > cg.line {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{byte(node.Position.Line - cg.line)}})
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

	// If statement
	case parser.NT_If:
		cg.generateIfStatement(node.Value.(*parser.IfNode))

	// Return
	case parser.NT_Return:
		// Generate returned expression
		if node.Value != nil {
			cg.generateExpression(node.Value.(*parser.Node), true)
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyRegAToD, NO_ARGS})
		}
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Return, NO_ARGS})

	default:
		panic("Unkown node.")
	}
}

func (cg *CodeGenerator) generateIfStatement(ifStatement *parser.IfNode) {
	jumpInstructions := make([]*VM.Instruction, len(ifStatement.IfStatements))
	jumpInstructionPositions := make([]int, len(ifStatement.IfStatements))

	// Generate else if conditions and jumps
	for i, statement := range ifStatement.IfStatements {
		// Generate condition expression
		cg.generateExpression(statement.Condition, true)
		// Generate conditional jump
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpIfTrue, []byte{0}})

		// Store the jump instruction and it's position so destination position can be set later
		jumpInstructions[i] = &cg.instructions[len(cg.instructions)-1]
		jumpInstructionPositions[i] = len(cg.instructions)
	}

	// Generate else body
	var jumpFromElse *VM.Instruction = nil
	if ifStatement.ElseBody != nil {
		// Generate body
		cg.generateStatements(ifStatement.ElseBody.Value.(*parser.ScopeNode))

		// Generate jump instruction that will jump over all elifs
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})
		// Store the jump instruction so it's destination can  be set later
		jumpFromElse = &cg.instructions[len(cg.instructions)-1]
	}
	// Store jump position
	jumpFromElsePosition := len(cg.instructions)

	// Generate else if bodies
	for i, statement := range ifStatement.IfStatements {
		// Set elif's conditional jump destination to next instruction
		jumpInstructions[i].InstructionValue[0] = byte(len(cg.instructions) - jumpInstructionPositions[i])
		// Generate elif's body
		cg.generateStatements(statement.Body.Value.(*parser.ScopeNode))

		// Generate jump instruction to the end of elif bodies
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})
		// Store it and it's position in the same array replacing the previous values
		jumpInstructions[i] = &cg.instructions[len(cg.instructions)-1]
		jumpInstructionPositions[i] = len(cg.instructions)
	}

	// Calculate distance from end of each of if/elif body to the end. Assign it to the jump instructions.
	endPosition := len(cg.instructions)
	for i, instruction := range jumpInstructions {
		instruction.InstructionValue[0] = byte(endPosition - jumpInstructionPositions[i])
	}

	// Assign distance to the end to the jump instruction in else block
	if jumpFromElse != nil {
		jumpFromElse.InstructionValue[0] = byte(endPosition - jumpFromElsePosition)
	}
}

func (cg *CodeGenerator) generateBody(functionNode *parser.FunctionDeclareNode) {
	cg.generateStatements(functionNode.Body.Value.(*parser.ScopeNode))

	if functionNode.Identifier == "entry" {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{0}})
	}
}

func (cg *CodeGenerator) generateStatements(scopeNode *parser.ScopeNode) {
	for _, node := range scopeNode.Statements {
		cg.generateNode(node)
	}
}

func (cg *CodeGenerator) generateFunctionCall(node *parser.Node) {
	// Generate arguments
	functionCall := node.Value.(*parser.FunctionCallNode)
	cg.generateArguments(functionCall.Arguments)

	// Call user defined function
	if functionCall.Number != -1 {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Call, []byte{byte(functionCall.Number)}})
		return
	}

	// Built-in function
	identifier := functionCall.Identifier

	// Check for overloaded built-in function
	_, overloaded := overloadedBuiltInFunctions[identifier]
	if overloaded {
		// Add parameter types to identifier so it can be matched to correct function
		for _, argumentType := range functionCall.ArgumentTypes {
			identifier = fmt.Sprintf("%s.%s", identifier, argumentType)
		}
	}

	// Try to look up built-in function
	builtInFunction, exists := builtInFunctions[identifier]

	// Function exists
	if exists {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunc, InstructionValue: []byte{builtInFunction}})
		// Function is exit()
	} else if functionCall.Identifier == "exit" {
		// Rewrite is as halt instruction
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{byte(functionCall.Arguments[0].Value.(*parser.LiteralNode).Value.(int64))}})
		// Normal function
	} else {
		panic("Unkown function.")
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
		// Add strings
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
