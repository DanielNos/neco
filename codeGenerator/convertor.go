package codeGenerator

import (
	"neko/parser"
	VM "neko/virtualMachine"
)

var nodeTypeToInstructionTypeInt = map[parser.NodeType]byte{
	parser.NT_Add:      VM.IT_IntAdd,
	parser.NT_Subtract: VM.IT_IntSubtract,
	parser.NT_Multiply: VM.IT_IntMultiply,
	parser.NT_Divide:   VM.IT_IntDivide,
	parser.NT_Power:    VM.IT_IntPower,
	parser.NT_Modulo:   VM.IT_IntModulo,
}

func (cg *CodeGenerator) lineToByte(line uint) byte {
	if line > 128 {
		cg.newError("There can't be more that 128 empty lines in succession.")
	}

	return byte(line) | 1<<7
}
