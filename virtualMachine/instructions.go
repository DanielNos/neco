package virtualMachine

const (
	IT_LoadConstant byte = iota
	IT_CallBuiltInFunction
	IT_Halt
	IT_IntAdd
	IT_IntSubtract
	IT_IntMultiply
	IT_IntDivide
	IT_IntPower
	IT_IntModulo
	IT_Push
)

var InstructionTypeToString = map[byte]string{
	IT_LoadConstant:        "LOAD_CONSTANT",
	IT_CallBuiltInFunction: "CALL_BUILTIN_FUNCTION",
	IT_Halt:                "HALT",
	IT_IntAdd:              "INT_ADD",
	IT_IntSubtract:         "INT_SUBTRACT",
	IT_IntMultiply:         "INT_MULTIPLY",
	IT_IntDivide:           "INT_DIVIDE",
	IT_IntPower:            "INT_POWER",
	IT_IntModulo:           "INT_MODULO",
	IT_Push:                "PUSH",
}

type Instruction struct {
	InstructionType  byte
	InstructionValue []byte
}
