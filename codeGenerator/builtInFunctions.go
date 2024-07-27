package codeGenerator

import VM "github.com/DanielNos/neco/virtualMachine"

var builtInFunctions = map[string]byte{
	"print":     VM.BIF_Print,
	"printLine": VM.BIF_PrintLine,

	"str": VM.BIF_AnyToString,

	"int.Bool": VM.BIF_BoolToInt,
	"int.Enum": VM.BIF_EnumToInt,

	"flt": VM.BIF_IntToFloat,

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

	"length": VM.BIF_StringLength,
	"size":   VM.BIF_ListLength,

	"toLower": VM.BIF_ToLower,
	"toUpper": VM.BIF_ToUpper,

	"randomInt":      VM.BIF_RandomInt,
	"randomFlt":      VM.BIF_RandomFloat,
	"randomRangeInt": VM.BIF_RandomRangeInt,

	"parseInt": VM.BIF_ParseInt,
	"parseFlt": VM.BIF_ParseFloat,

	"trace": VM.BIF_Trace,
	"panic": VM.BIF_Panic,
}

var overloadedBuiltInFunctions = map[string]struct{}{
	"int": {},
}
