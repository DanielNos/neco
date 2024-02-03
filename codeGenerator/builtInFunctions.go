package codeGenerator

import VM "neco/virtualMachine"

var builtInFunctions = map[string]byte{
	"print":     VM.BIF_Print,
	"printLine": VM.BIF_PrintLine,

	"str.Bool":  VM.BIF_BoolToString,
	"str.Int":   VM.BIF_IntToString,
	"str.Float": VM.BIF_FloatToString,

	"int.Bool": VM.BIF_BoolToInt,
	"flt.Int":  VM.BIF_IntToFloat,

	"floor":      VM.BIF_Floor,
	"floorToInt": VM.BIF_FloorToInt,
	"ceil":       VM.BIF_Ceil,
	"ceilToInt":  VM.BIF_CeilToInt,
	"round":      VM.BIF_Round,
	"roundToInt": VM.BIF_RoundToInt,
	"absInt":     VM.BIF_AbsInt,
	"absFlt":     VM.BIF_AbsFloat,

	"readLine": VM.BIF_ReadLine,
	"readChar": VM.BIF_ReadChar,

	"length":  VM.BIF_Length,
	"toLower": VM.BIF_ToLower,
	"toUpper": VM.BIF_ToUpper,

	"randomInt":      VM.BIF_RandomInt,
	"randomFlt":      VM.BIF_RandomFloat,
	"randomRangeInt": VM.BIF_RandomRangeInt,
}

var overloadedBuiltInFunctions = map[string]struct{}{
	"int": {},
	"flt": {},
	"str": {},
}
