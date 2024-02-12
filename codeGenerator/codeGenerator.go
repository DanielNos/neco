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
var ARGS_ZERO = []byte{0}

type CodeGenerator struct {
	filePath string
	tree     *parser.Node
	optimize bool

	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int
	constants       []interface{} // int64/float64/string

	instructions []VM.Instruction

	variableIdentifierCounters *dataStructures.Stack // of uint8
	variableIdentifiers        *dataStructures.Stack // of map[string]uint8

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
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, ARGS_ZERO})

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
	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	// Variable declaration
	case parser.NT_VariableDeclare:
		cg.generateVariableDeclaration(node)

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

	// Scope
	case parser.NT_Scope:
		cg.enterScope()

		cg.generateStatements(node.Value.(*parser.ScopeNode))

		cg.leaveScope()

	// Loops
	case parser.NT_Loop:
		cg.enterScope()

		startPosition := len(cg.instructions)
		cg.generateStatements(node.Value.(*parser.Node).Value.(*parser.ScopeNode))
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_JumpBack, []byte{byte(len(cg.instructions) - startPosition)}})

		cg.leaveScope()

	default:
		panic("Unkown node.")
	}
}

func (cg *CodeGenerator) generateStatements(scopeNode *parser.ScopeNode) {
	for _, node := range scopeNode.Statements {
		cg.generateNode(node)
	}
}

func (cg *CodeGenerator) generateVariableDeclaration(node *parser.Node) {
	variable := node.Value.(*parser.VariableDeclareNode)

	for i := 0; i < len(variable.Identifiers); i++ {
		cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]] = cg.variableIdentifierCounters.Top.Value.(uint8)

		switch variable.VariableType.DataType {
		case parser.DT_Bool:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareBool, []byte{cg.variableIdentifierCounters.Top.Value.(uint8)}})
		case parser.DT_Int:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareInt, []byte{cg.variableIdentifierCounters.Top.Value.(uint8)}})
		case parser.DT_Float:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareFloat, []byte{cg.variableIdentifierCounters.Top.Value.(uint8)}})
		case parser.DT_String:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareString, []byte{cg.variableIdentifierCounters.Top.Value.(uint8)}})
		}

		cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1
	}
}

func (cg *CodeGenerator) enterScope() {
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PushScopeUnnamed, NO_ARGS})
	cg.variableIdentifierCounters.Push(cg.variableIdentifierCounters.Top.Value)
	cg.variableIdentifiers.Push(map[string]uint8{})
}

func (cg *CodeGenerator) leaveScope() {
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PopScope, NO_ARGS})
	cg.variableIdentifierCounters.Pop()
	cg.variableIdentifiers.Pop()
}
