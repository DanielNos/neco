package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateFunction(function *parser.FunctionDeclareNode) {
	// Store start position
	cg.functions = append(cg.functions, len(cg.instructions))

	// Push scope
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PushScope, []byte{byte(cg.stringConstants[function.Identifier])}})
	cg.variableIdentifierCounters.Push(uint8(0))
	cg.variableIdentifiers.Push(map[string]uint8{})

	// Pop arguments and store them as variables
	parameterCount := len(function.Parameters)
	for i := parameterCount - 1; i >= 0; i-- {
		// Declare variable for argument
		switch function.Parameters[i].DataType.DataType {
		case parser.DT_Bool:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareBool, NO_ARGS})
		case parser.DT_Int:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareInt, NO_ARGS})
		case parser.DT_Float:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareFloat, NO_ARGS})
		case parser.DT_String:
			cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_DeclareString, NO_ARGS})
		}

		// Store argument in variable
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PopArgStackVariable, []byte{byte(parameterCount - i)}})
	}

	// Generate function body
	cg.generateBody(function)

	// Leave scope
	cg.variableIdentifierCounters.Pop()
	cg.variableIdentifiers.Pop()

	// Return
	cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_Return, NO_ARGS})
}
