package codeGenerator

import (
	"encoding/binary"
	data "neco/dataStructures"
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

var comparisonOperatorToIntInstruction = map[parser.NodeType]byte{
	parser.NT_Lower:        VM.IT_IntLower,
	parser.NT_Greater:      VM.IT_IntGreater,
	parser.NT_LowerEqual:   VM.IT_IntLowerEqual,
	parser.NT_GreaterEqual: VM.IT_IntGreaterEqual,
}

var comparisonOperatorToFloatInstruction = map[parser.NodeType]byte{
	parser.NT_Lower:        VM.IT_FloatLower,
	parser.NT_Greater:      VM.IT_FloatGreater,
	parser.NT_LowerEqual:   VM.IT_FloatLowerEqual,
	parser.NT_GreaterEqual: VM.IT_FloatGreaterEqual,
}

var dataTypeToDeclareInstruction = map[data.DType]byte{
	data.DT_Bool:   VM.IT_DeclareBool,
	data.DT_Int:    VM.IT_DeclareInt,
	data.DT_Float:  VM.IT_DeclareFloat,
	data.DT_String: VM.IT_DeclareString,
	data.DT_List:   VM.IT_DeclareList,
	data.DT_Set:    VM.IT_DeclareSet,
	data.DT_Struct: VM.IT_DeclareObject,
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
