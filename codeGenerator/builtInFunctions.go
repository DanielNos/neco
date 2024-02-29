package codeGenerator

import VM "neco/virtualMachine"

var builtInFunctions = map[string]byte{
	"print":     VM.BIF_Print,
	"printLine": VM.BIF_PrintLine,

	"str": VM.BIF_AnyToString,

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

	"length.String": VM.BIF_StringLength,

	"toLower": VM.BIF_ToLower,
	"toUpper": VM.BIF_ToUpper,

	"randomInt":      VM.BIF_RandomInt,
	"randomFlt":      VM.BIF_RandomFloat,
	"randomRangeInt": VM.BIF_RandomRangeInt,

	"trace": VM.BIF_Trace,
}

var overloadedBuiltInFunctions = map[string]struct{}{
	"int":    {},
	"flt":    {},
	"length": {},
}
