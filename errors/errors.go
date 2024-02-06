package errors

const MAX_ERROR_COUNT = 15

const (
	INVALID_FLAGS = iota + 1
	LEXICAL
	SYNTAX
	SEMANTIC
	READ_PROGRAM
	INCOMPATIBLE_VERSION
	CODE_GENERATION
	UNKNOWN_INSTRUCTION
	STACK_OVERFLOW
)
