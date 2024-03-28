package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

func (cg *CodeGenerator) generateFunction(functionNode *parser.Node) {
	function := functionNode.Value.(*parser.FunctionDeclareNode)

	// Store start position
	cg.functions = append(cg.functions, len(*cg.target))

	// Push scope
	cg.enterScope(&function.Identifier)

	// Pop arguments and store them as variables
	parameterCount := len(function.Parameters)
	for i := parameterCount - 1; i >= 0; i-- {

		// Declare variable for argument
		id := cg.variableIdentifierCounters.Top.Value.(uint8)
		cg.variableIdentifiers.Top.Value.(map[string]uint8)[function.Parameters[i].Identifier] = id

		// Generate declaration instruction
		cg.generateVariableDeclarator(function.Parameters[i].DataType, &id)

		cg.variableIdentifierCounters.Top.Value = cg.variableIdentifierCounters.Top.Value.(uint8) + 1

		// Store argument from stack in the variable
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_StoreAndPop, []byte{id}})
	}

	// Generate function body
	cg.generateStatements(function.Body.Value.(*parser.ScopeNode))

	// Generate line offset of closing brace
	if cg.line < functionNode.Position.EndLine {
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_LineOffset, []byte{byte(functionNode.Position.EndLine - cg.line)}})
		cg.line = functionNode.Position.EndLine
	}

	// Return
	*cg.target = append(*cg.target, VM.Instruction{VM.IT_Return, NO_ARGS})
}

func (cg *CodeGenerator) generateFunctionCall(node *parser.Node) {
	// Generate arguments
	functionCall := node.Value.(*parser.FunctionCallNode)
	cg.generateArguments(functionCall.Arguments)

	// Call user defined function
	if functionCall.Number != -1 {
		*cg.target = append(*cg.target, VM.Instruction{VM.IT_Call, []byte{byte(functionCall.Number)}})
		return
	}

	// Built-in function
	identifier := functionCall.Identifier

	// Check for overloaded built-in function
	_, overloaded := overloadedBuiltInFunctions[identifier]
	if overloaded {
		// Add parameter types to identifier so it can be matched to correct function
		for _, argumentType := range functionCall.ArgumentTypes {
			identifier = identifier + "." + argumentType.Signature()
		}
	}

	// Try to look up built-in function
	builtInFunction, exists := builtInFunctions[identifier]

	// It's a built-in function
	if exists {
		*cg.target = append(*cg.target, VM.Instruction{InstructionType: VM.IT_CallBuiltInFunc, InstructionValue: []byte{builtInFunction}})
		// Function is exit()
	} else if functionCall.Identifier == "exit" {
		// Convert exit function to halt instruction
		*cg.target = append(*cg.target, VM.Instruction{InstructionType: VM.IT_Halt, InstructionValue: []byte{byte(functionCall.Arguments[0].Value.(*parser.LiteralNode).Value.(int64))}})
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
