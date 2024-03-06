package codeGenerator

import (
	"fmt"
	"neco/parser"
	VM "neco/virtualMachine"
	"strings"
)

func (cg *CodeGenerator) generateFunction(functionNode *parser.Node) {
	function := functionNode.Value.(*parser.FunctionDeclareNode)

	// Store start position
	cg.functions = append(cg.functions, len(cg.instructions))

	// Push scope
	cg.enterScope()

	// Pop arguments and store them as variables
	parameterCount := len(function.Parameters)
	for i := parameterCount - 1; i >= 0; i-- {

		// Declare variable for argument
		identifier := cg.variableIdentifierCounters.Top.Value.(uint8)
		cg.variableIdentifiers.Top.Value.(map[string]uint8)[function.Parameters[i].Identifier] = identifier

		// Generate declaration instruction
		cg.generateVariableDeclarator(function.Parameters[i].DataType, true)

		cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1

		// Store argument from stack in the variable
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Store, []byte{identifier}})
	}

	// Generate function body
	cg.generateStatements(function.Body.Value.(*parser.ScopeNode))

	// Leave scope
	cg.leaveScope()

	// Generate line offset of closing brace
	if cg.line < functionNode.Position.EndLine {
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_LineOffset, []byte{byte(functionNode.Position.EndLine - cg.line)}})
		cg.line = functionNode.Position.EndLine
	}

	// Return
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Return, NO_ARGS})
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
			identifier = fmt.Sprintf("%s.%s", identifier, argumentType.Signature())
		}
	}

	// Try to look up built-in function
	builtInFunction, exists := builtInFunctions[identifier]

	// It's a built-in function
	if exists {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunc, InstructionValue: []byte{builtInFunction}})
		// Function is exit()
	} else if functionCall.Identifier == "exit" {
		// Convert exit function to halt instruction
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{byte(functionCall.Arguments[0].Value.(*parser.LiteralNode).Value.(int64))}})
		// List length
	} else if strings.HasPrefix(identifier, "length") {
		cg.instructions = append(cg.instructions, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunc, InstructionValue: []byte{VM.BIF_ListLength}})
		// Unknown function
	} else {
		panic("Unkown function.")
	}
}

func (cg *CodeGenerator) generateArguments(arguments []*parser.Node) {
	for _, argument := range arguments {
		cg.generateExpression(argument)
	}
}
