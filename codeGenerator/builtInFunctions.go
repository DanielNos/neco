package codeGenerator

import VM "neco/virtualMachine"

var builtInFunctions = map[string]byte{
	"print":     VM.BIF_Print,
	"printLine": VM.BIF_PrintLine,

	"boolToStr": VM.BIF_BoolToString,
	"intToStr":  VM.BIF_IntToString,
	"fltToStr":  VM.BIF_FloatToString,

	"boolToInt": VM.BIF_BoolToInt,
	"intToFlt":  VM.BIF_IntToFloat,

	"floor":      VM.BIF_Floor,
	"floorToInt": VM.BIF_FloorToInt,
	"ceil":       VM.BIF_Ceil,
	"ceilToInt":  VM.BIF_CeilToInt,
	"round":      VM.BIF_Round,
	"roundToInt": VM.BIF_RoundToInt,

	"readLine": VM.BIF_ReadLine,
	"readChar": VM.BIF_ReadChar,
}
