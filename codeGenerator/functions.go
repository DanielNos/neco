package codeGenerator

import (
	"fmt"
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
		switch function.Parameters[i].DataType.DType {
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
		cg.instructions = append(cg.instructions, VM.Instruction{VM.IT_PopArgStackToVariable, []byte{byte(parameterCount - i)}})
	}

	// Generate function body
	cg.generateStatements(function.Body.Value.(*parser.ScopeNode))

	// Leave scope
	cg.variableIdentifierCounters.Pop()
	cg.variableIdentifiers.Pop()

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
			identifier = fmt.Sprintf("%s.%s", identifier, argumentType)
		}
	}

	// Try to look up built-in function
	builtInFunction, exists := builtInFunctions[identifier]

	// It's a built-in function
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
