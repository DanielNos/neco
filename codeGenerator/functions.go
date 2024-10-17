package codeGenerator

import (
	"github.com/DanielNos/neco/parser"
	VM "github.com/DanielNos/neco/virtualMachine"
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
		id := cg.scopes.Top.Value.(*Scope).variableIdentifierCounter
		cg.scopes.Top.Value.(*Scope).variableIdentifiers[function.Parameters[i].Identifier] = id

		// Generate declaration instruction
		cg.generateVariableDeclarator(function.Parameters[i].DataType, &id)

		cg.scopes.Top.Value.(*Scope).variableIdentifierCounter++

		// Store argument from stack in the variable
		cg.addInstruction(VM.IT_StoreAndPop, id)
	}

	// Generate function body
	cg.generateStatements(function.Body.Value.(*parser.ScopeNode))

	// Generate line offset of closing brace
	if cg.currentLine(functionNode) < functionNode.Position.EndLine {
		cg.addInstruction(VM.IT_LineOffset, byte(functionNode.Position.EndLine-cg.currentLine(functionNode)))
		cg.currentLines[*functionNode.Position.File] = int(functionNode.Position.EndLine)
	}

	// Return
	cg.addInstruction(VM.IT_Return)
}

func (cg *CodeGenerator) generateFunctionCall(node *parser.Node) {
	// Generate arguments
	functionCall := node.Value.(*parser.FunctionCallNode)
	cg.generateArguments(functionCall.Arguments)

	// Call user defined function
	if functionCall.Number != -1 {
		cg.addInstruction(VM.IT_Call, byte(functionCall.Number))
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
		cg.addInstruction(VM.IT_CallBuiltInFunc, builtInFunction)
		// Function is exit()
	} else if functionCall.Identifier == "exit" {
		// Convert exit function to halt instruction
		cg.addInstruction(VM.IT_Halt, byte(functionCall.Arguments[0].Value.(*parser.LiteralNode).Value.(int64)))
		// Unknown function
	} else {
		panic("Unknown function.")
	}
}

func (cg *CodeGenerator) generateArguments(arguments []*parser.Node) {
	for _, argument := range arguments {
		cg.generateExpression(argument)
	}
}
