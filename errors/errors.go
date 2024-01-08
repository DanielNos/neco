package errors

const MAX_ERROR_COUNT = 15

const (
	ERROR_INTERNAL = iota + 1
	ERROR_INVALID_USE
	ERROR_LEXICAL
	ERROR_SYNTAX
	ERROR_SEMANTIC
	ERROR_READ_PROGRAM
	ERROR_INCOMPATIBLE_VERSION
)
