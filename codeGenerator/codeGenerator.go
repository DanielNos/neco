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

const MAX_UINT8 = 255
const IGNORE_INSTRUCTION byte = 255

type Break struct {
	instruction         *VM.Instruction
	instructionPosition int
}

type ScopeType uint8

const (
	ST_Root ScopeType = iota
	ST_Unnamed
	ST_Function
)

type Scope struct {
	scopeType                 ScopeType
	variableIdentifierCounter uint8
	variableIdentifiers       map[string]uint8
}

type CodeGenerator struct {
	tree     *parser.Node
	optimize bool

	intConstants    map[int64]int
	floatConstants  map[float64]int
	stringConstants map[string]int
	Constants       []any // int64/float64/string

	FirstLine             uint
	GlobalsInstructions   []VM.Instruction
	FunctionsInstructions []VM.Instruction
	target                *[]VM.Instruction

	functions []int // Function number : function start

	scopeBreaks     *data.Stack // break
	loopScopeDepths *data.Stack // int

	scopes *data.Stack // Scope

	line uint

	ErrorCount int
}

func NewGenerator(tree *parser.Node, intConstants map[int64]int, floatConstants map[float64]int, stringConstants map[string]int, optimize bool) *CodeGenerator {
	codeGenerator := &CodeGenerator{
		tree:     tree,
		optimize: optimize,

		intConstants:    intConstants,
		floatConstants:  floatConstants,
		stringConstants: stringConstants,
		Constants:       make([]any, len(intConstants)+len(floatConstants)+len(stringConstants)),

		GlobalsInstructions:   []VM.Instruction{},
		FunctionsInstructions: []VM.Instruction{},

		functions: []int{},

		scopeBreaks:     data.NewStack(),
		loopScopeDepths: data.NewStack(),

		scopes: data.NewStack(),

		line:       0,
		ErrorCount: 0,
	}

	// Enter root scope
	codeGenerator.pushScope(ST_Root)

	return codeGenerator
}

func (cg *CodeGenerator) Generate() {
	// Generate constant IDs
	cg.generateConstantIDs()

	// Get root statement list
	statements := cg.tree.Value.(*parser.ModuleNode).Statements.Statements

	// No instructions, generate line offset and halt instruction
	if len(statements) == 0 {
		cg.FunctionsInstructions = append(cg.FunctionsInstructions, VM.Instruction{VM.IT_LineOffset, []byte{1}})

		logger.Warning("Source code doesn't contain any statements. No instructions will be generated.")
		return
	}

	// Store first line
	cg.FirstLine = statements[0].Position.StartLine

	// Generate code
	cg.generateGlobals(statements)

	// Generate call to entry function
	cg.generateFunctions(statements)

	// Optimize instructions
	if cg.optimize {
		codeOptimizer.Optimize(&cg.GlobalsInstructions)
		codeOptimizer.Optimize(&cg.FunctionsInstructions)
	}

	if len(cg.GlobalsInstructions) == 0 && len(cg.FunctionsInstructions) == 0 {
		logger.Warning("No instructions were generated. Binary will be empty.")
	}
}

func (cg *CodeGenerator) addInstruction(instructionType byte, instructionArguments ...byte) {
	(*cg.target) = append(*cg.target, VM.Instruction{instructionType, instructionArguments})
}

func (cg *CodeGenerator) generateGlobals(statements []*parser.Node) {
	// Reset line
	cg.line = cg.FirstLine
	cg.target = &cg.GlobalsInstructions

	for _, node := range statements {
		// Generate line offset if line changed
		if cg.line < node.Position.StartLine {
			cg.addInstruction(VM.IT_LineOffset, byte(node.Position.StartLine-cg.line))
			cg.line = node.Position.StartLine
		}

		// Generate only declarations and assignments
		if node.NodeType == parser.NT_VariableDeclaration {
			cg.generateVariableDeclaration(node)
		} else if node.NodeType == parser.NT_Assign {
			cg.generateAssignment(node.Value.(*parser.AssignNode))
		} else {
			// Stop if something else is found (globals are only at the start of the tree)
			break
		}
	}
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
		cg.Constants[id] = key
		cg.stringConstants[key] = id
		id++
	}

	// Integers
	for key := range cg.intConstants {
		cg.Constants[id] = key
		cg.intConstants[key] = id
		id++
	}

	// Floats
	for key := range cg.floatConstants {
		cg.Constants[id] = key
		cg.floatConstants[key] = id
		id++
	}

	// More than 255 constants
	if id >= math.MaxUint8 {
		logger.Error(fmt.Sprintf("Constant pool overflow with %d constants. Constant pool can only contain maximum of %d constants.", id, math.MaxUint8))
	}
}

func (cg *CodeGenerator) generateFunctions(statements []*parser.Node) {
	// Reset line
	cg.line = cg.FirstLine
	cg.target = &cg.FunctionsInstructions

	cg.FunctionsInstructions = append(cg.FunctionsInstructions, VM.Instruction{IGNORE_INSTRUCTION, []byte{}})

	for _, node := range statements {
		if node.NodeType == parser.NT_FunctionDeclaration {
			// Generate line offset if line changed
			if cg.line < node.Position.StartLine {
				cg.addInstruction(VM.IT_LineOffset, byte(node.Position.StartLine-cg.line))
				cg.line = node.Position.StartLine
			}

			// Set first function call function id
			if node.Value.(*parser.FunctionDeclareNode).Identifier == "entry" {
				cg.FunctionsInstructions[0].InstructionType = VM.IT_Call
				cg.FunctionsInstructions[0].InstructionValue = append(cg.FunctionsInstructions[0].InstructionValue, byte(len(cg.functions)))
			}

			cg.generateFunction(node)
		}
	}
}

func (cg *CodeGenerator) generateNode(node *parser.Node) {
	// If node line has changed, insert line offset instruction
	if node.Position.StartLine > cg.line {
		cg.addInstruction(VM.IT_LineOffset, byte(node.Position.StartLine-cg.line))
		cg.line = node.Position.StartLine
	}

	switch node.NodeType {
	// Function call
	case parser.NT_FunctionCall:
		cg.generateFunctionCall(node)

	// Variable declaration
	case parser.NT_VariableDeclaration:
		cg.generateVariableDeclaration(node)

	// Assignment
	case parser.NT_Assign:
		cg.generateAssignment(node.Value.(*parser.AssignNode))

	// If statement
	case parser.NT_If:
		cg.generateIfStatement(node.Value.(*parser.IfNode))

	// Return
	case parser.NT_Return:
		// Generate returned expression
		if node.Value != nil {
			cg.generateExpression(node.Value.(*parser.Node))
		}

		// Drop unnamed scopes
		currentScope := cg.scopes.Top
		for currentScope.Value.(*Scope).scopeType != ST_Function {
			cg.addInstruction(VM.IT_PopScope)
			currentScope = currentScope.Previous
		}

		cg.addInstruction(VM.IT_Return)

	// Scope
	case parser.NT_Scope:
		cg.generateScope(node.Value.(*parser.ScopeNode), nil)

	// Loops
	case parser.NT_Loop:
		cg.generateLoop(node.Value.(*parser.Node))

	case parser.NT_ForLoop:
		cg.generateForLoop(node)

	// Break
	case parser.NT_Break:
		// Generate scope drops
		for i := 0; i < cg.scopes.Size-cg.loopScopeDepths.Top.Value.(int)+1; i++ {
			cg.addInstruction(VM.IT_PopScope)
		}

		// Generate jump
		cg.addInstruction(VM.IT_Jump, 0)

		// Store it so it's destination can be set at the end of the loop
		cg.scopeBreaks.Top.Value = append(cg.scopeBreaks.Top.Value.([]Break), Break{&(*cg.target)[len(*cg.target)-1], len(*cg.target)})

	case parser.NT_ListAssign:
		// Generate index expression
		cg.generateExpression(node.Value.(*parser.ListAssignNode).IndexExpression)

		// Generate assigned expression
		cg.generateExpression(node.Value.(*parser.ListAssignNode).AssignedExpression)

		// Generate list assign instruction
		cg.addInstruction(VM.IT_SetListAtPrevToCurr, cg.findVariableIdentifier(node.Value.(*parser.ListAssignNode).Identifier))

	// Delete
	case parser.NT_Delete:
		cg.generateDeletion(node.Value.(*parser.Node))

	// Match
	case parser.NT_Match:
		cg.generateMatch(node.Value.(*parser.MatchNode), false)

	default:
		panic("Unknown node " + parser.NodeTypeToString[node.NodeType])
	}
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
	// Look for variable in current scope
	currentNode := cg.scopes.Top
	id, found := currentNode.Value.(*Scope).variableIdentifiers[identifier]

	// Find variable in lower scopes
	for !found && currentNode != nil {
		// Move to previous node and try to find variable
		currentNode = currentNode.Previous
		id, found = currentNode.Value.(*Scope).variableIdentifiers[identifier]
	}

	// Couldn't find variable
	if !found {
		panic("Failed to find variable id: " + identifier + ".")
	}

	return id
}
