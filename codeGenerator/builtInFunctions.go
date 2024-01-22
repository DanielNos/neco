package codeGenerator

import VM "neco/virtualMachine"

var builtInFunctions = map[string]byte{
	"print":     VM.BIF_Print,
	"printLine": VM.BIF_PrintLine,
	"bool2str":  VM.BIF_Bool2String,
	"int2str":   VM.BIF_Int2String,
	"flt2str":   VM.BIF_Float2String,
}
