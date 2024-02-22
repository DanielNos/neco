package codeGenerator

import (
	"fmt"
	"math"
	"neco/codeOptimizer"
	data "neco/dataStructures"
	"neco/errors"
	"neco/logger"
	"neco/parser"
	VM "neco/virtualMachine"
)

var NO_ARGS = []byte{}

const MAX_UINT8 = 255

type Break struct {
	instruction         *VM.Instruction
	instructionPosition int
}

type CodeGenerator struct {
	filePath string
	tree     *parser.Node
	optimize bool

	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int
	constants       []interface{} // int64/float64/string

	instructions []VM.Instruction

	variableIdentifierCounters *data.Stack // of uint8
	variableIdentifiers        *data.Stack // of map[string]uint8

	functions []int // Function number : function start

	scopeBreaks     *data.Stack // break
	loopScopeDepths *data.Stack // int

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

		data.NewStack(),
		data.NewStack(),

		[]int{},

		data.NewStack(),
		data.NewStack(),

		0,
		0,
	}

	if !optimize {
		logger.Warning("Byte code optimization disabled.")
	}

	// Enter root scope
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
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{1}})

		logger.Warning("Source code doesn't contain any symbols. Binary will be generated, but will contain no instructions.")

		return &cg.instructions
	}

	// Generate functions
	cg.line = statements[0].Position.Line
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{byte(cg.line)}})

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

		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_StoreOpA, []byte{cg.findVariableIdentifier(assignNode.Identifier)}})

	// If statement
	case parser.NT_If:
		cg.generateIfStatement(node.Value.(*parser.IfNode))

	// Return
	case parser.NT_Return:
		// Generate returned expression
		if node.Value != nil {
			cg.generateExpression(node.Value.(*parser.Node), true)
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpAToReturn, NO_ARGS})
		}
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Return, NO_ARGS})

	// Scope
	case parser.NT_Scope:
		cg.generateScope(node.Value.(*parser.ScopeNode))

	// Loops
	case parser.NT_Loop:
		cg.generateLoop(node.Value.(*parser.Node))

	// Break
	case parser.NT_Break:
		// Generate scope drops
		for i := 0; i < cg.variableIdentifiers.Size-cg.loopScopeDepths.Top.Value.(int); i++ {
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PopScope, NO_ARGS})
		}

		// Generate jump
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Jump, []byte{0}})

		// Store it so it's destination can be set at the end of the loop
		cg.scopeBreaks.Top.Value = append(cg.scopeBreaks.Top.Value.([]Break), Break{&cg.instructions[len(cg.instructions)-1], len(cg.instructions)})

	case parser.NT_ListAssign:
		// Generate assigned expression and store it in reg E
		cg.generateExpression(node.Value.(*parser.ListAssignNode).AssignedExpression, true)
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyOpAToListA, NO_ARGS})

		// Generate index expression
		cg.generateExpression(node.Value.(*parser.ListAssignNode).IndexExpression, true)

		// Move assigned expression from reg E to reg B
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_CopyListAToOpB, NO_ARGS})

		// Generate list assign instruction
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_SetListAtAToB, []byte{cg.findVariableIdentifier(node.Value.(*parser.ListAssignNode).Identifier)}})

	default:
		panic("Unkown node.")
	}
}

func (cg *CodeGenerator) generateVariableDeclaration(node *parser.Node) {
	variable := node.Value.(*parser.VariableDeclareNode)

	for i := 0; i < len(variable.Identifiers); i++ {
		cg.variableIdentifiers.Top.Value.(map[string]uint8)[variable.Identifiers[i]] = cg.variableIdentifierCounters.Top.Value.(uint8)

		cg.generateVariableDeclarator(variable.DataType, true)

		cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1
	}
}

func (cg *CodeGenerator) generateVariableDeclarator(dataType parser.DataType, passId bool) {
	args := NO_ARGS
	if passId {
		args = []byte{cg.variableIdentifierCounters.Top.Value.(uint8)}
	}

	switch dataType.DType {
	case parser.DT_Bool:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareBool, args})
	case parser.DT_Int:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareInt, args})
	case parser.DT_Float:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareFloat, args})
	case parser.DT_String:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareString, args})
	case parser.DT_List:
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareList, args})

		subType := dataType.SubType
		for subType != nil {
			cg.generateVariableDeclarator(subType.(parser.DataType), false)
			subType = subType.(parser.DataType).SubType
		}
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

func (cg *CodeGenerator) generateStatements(scopeNode *parser.ScopeNode) {
	for _, node := range scopeNode.Statements {
		cg.generateNode(node)
	}
}

func (cg *CodeGenerator) generateScope(scopeNode *parser.ScopeNode) {
	cg.enterScope()
	cg.generateStatements(scopeNode)
	cg.leaveScope()
}

func updateJumpDistance(instruction *VM.Instruction, distance int, extendedInstructionType byte) {
	// If distance is larger than 255, change instruction type to extended jump
	if distance > MAX_UINT8 {
		instruction.InstructionType = extendedInstructionType
		instruction.InstructionValue = intTo2Bytes(distance)
	} else {
		instruction.InstructionValue[0] = byte(distance)
	}
}

func (cg *CodeGenerator) findVariableIdentifier(identifier string) uint8 {
	// Look for vairable in current scope
	currentNode := cg.variableIdentifiers.Top
	id, found := currentNode.Value.(map[string]uint8)[identifier]

	// Find variable in lower scopes
	for !found && currentNode != nil {
		// Move to previous node and try to find variable
		currentNode = currentNode.Previous
		id, found = currentNode.Value.(map[string]uint8)[identifier]
	}

	// Couldn't find variable
	if !found {
		panic(fmt.Sprintf("Failed to find variable id: %s.", identifier))
	}

	return id
}
