package codeGenerator

import (
	"encoding/binary"
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

var logicalOperatorToIntInstruction = map[parser.NodeType]byte{
	parser.NT_Lower:        VM.IT_LowerInt,
	parser.NT_Greater:      VM.IT_GreaterInt,
	parser.NT_LowerEqual:   VM.IT_LowerEqualInt,
	parser.NT_GreaterEqual: VM.IT_GreaterEqualInt,
}

var logicalOperatorToFloatInstruction = map[parser.NodeType]byte{
	parser.NT_Lower:        VM.IT_LowerFloat,
	parser.NT_Greater:      VM.IT_GreaterFloat,
	parser.NT_LowerEqual:   VM.IT_LowerEqualFloat,
	parser.NT_GreaterEqual: VM.IT_GreaterEqualFloat,
}

func (cg *CodeGenerator) lineToInstruction(line byte) byte {
	if line > 128 {
		cg.newError("There can't be more that 128 empty lines in succession.")
	}

	return line - 1 | 1<<7
}

func intTo2Bytes(int int) []byte {
	bytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytes, uint16(int))

	return bytes
}
