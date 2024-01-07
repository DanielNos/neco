package virtualMachine

const (
	IT_LoadConstant byte = iota
	IT_CallBuiltInFunction
	IT_Halt
)

var InstructionTypeToString = map[byte]string{
	IT_LoadConstant:        "LOAD_CONSTANT",
	IT_CallBuiltInFunction: "CALL_BUILTIN_FUNCTION",
	IT_Halt:                "HALT",
}

type InstructionValue interface{}

type Instruction struct {
	InstructionType byte

	ValueA InstructionValue
	ValueB InstructionValue
	ValueC InstructionValue
}
