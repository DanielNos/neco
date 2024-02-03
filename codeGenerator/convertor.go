package codeGenerator

import (
	"neco/parser"
	VM "neco/virtualMachine"
)

var intOperatorToInstruction = map[parser.NodeType]byte{
	parser.NT_Add:      VM.IT_IntAdd,
	parser.NT_Subtract: VM.IT_IntSubtract,
	parser.NT_Multiply: VM.IT_IntMultiply,
	parser.NT_Divide:   VM.IT_IntDivide,
	parser.NT_Power:    VM.IT_IntPower,
	parser.NT_Modulo:   VM.IT_IntModulo,
}

var floatOperatorToInstruction = map[parser.NodeType]byte{
	parser.NT_Add:      VM.IT_FloatAdd,
	parser.NT_Subtract: VM.IT_FloatSubtract,
	parser.NT_Multiply: VM.IT_FloatMultiply,
	parser.NT_Divide:   VM.IT_FloatDivide,
	parser.NT_Power:    VM.IT_FloatPower,
	parser.NT_Modulo:   VM.IT_FloatModulo,
}

func (cg *CodeGenerator) lineToInstruction(line byte) byte {
	if line > 128 {
		cg.newError("There can't be more that 128 empty lines in succession.")
	}

	return line - 1 | 1<<7
}
