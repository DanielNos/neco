package codeGenerator

import (
	"encoding/binary"
	"math"
	"neco/errors"
	"neco/logger"
	VM "neco/virtualMachine"
	"os"
)

var STRING_TERMINATOR = []byte{0}

const CONSTANTS_SEGMENT = 0
const (
	STRINGS_SEGMENT = 0
	INTS_SEGMENT    = 1
	FLOATS_SEGMENT  = 2
)

const CODE_SEGMENT = 1
const (
	METADATA_SEGMENT          = 0
	GLOBALS_SEGMENT           = 1
	FUNCTIONS_INDEXES_SEGMENT = 2
	FUNCTIONS_SEGMENT         = 3
)

type CodeWriter struct {
	codeGenerator *CodeGenerator
	file          *os.File
}

func NewCodeWriter(codeGenerator *CodeGenerator) *CodeWriter {
	return &CodeWriter{codeGenerator, nil}
}

func (cw *CodeWriter) Write() {
	file, _ := os.Create(cw.codeGenerator.filePath)
	cw.file = file

	file.WriteString("NECO")
	file.Write([]byte{0, VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH})

	cw.writeConstantsSegment()
	cw.writeCodeSegment()

	file.Close()
}

func (cw *CodeWriter) writeInstructions(instructions *[]VM.Instruction) {
	// Write instructions
	for _, instruction := range *instructions {
		// Skip instruction removed by code optimizer
		if instruction.InstructionType == 255 {
			continue
		}

		// Convert line offset instruction to single byte
		if instruction.InstructionType == VM.IT_LineOffset {
			cw.file.Write([]byte{cw.codeGenerator.lineToInstruction(instruction.InstructionValue[0])})
			continue
		}
		cw.file.Write([]byte{instruction.InstructionType}) // Write instruction
		cw.file.Write(instruction.InstructionValue)        // Write arguments
	}
}

func int64ToByte3(value int64) []byte {
	intBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(intBytes, uint64(value))
	return intBytes[5:8]
}

func (cw *CodeWriter) getFilePosition() int64 {
	info, _ := cw.file.Stat()
	return info.Size()
}

func (cw *CodeWriter) writeCodeSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("CODE")

	cw.writeMetaData()
	cw.writeGlobals()
	cw.writeFunctionIndexes()
	cw.writeFunctions()

	cw.file.WriteAt([]byte{CODE_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeMetaData() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("META")

	cw.file.Write([]byte{byte(cw.codeGenerator.FirstLine)}) // Write first line

	cw.file.WriteAt([]byte{METADATA_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeGlobals() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("GLOB")

	cw.writeInstructions(&cw.codeGenerator.GlobalsInstructions)

	cw.file.WriteAt([]byte{GLOBALS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeFunctionIndexes() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("FUNI")

	lastFunction := 0

	for _, function := range cw.codeGenerator.functions {
		// Can't have bigger difference than 1 byte
		if function-lastFunction > 256 {
			logger.Fatal(errors.CODE_GENERATION, "Function distance bigger than 256.")
		}

		// Write difference between this function and the last one
		cw.file.Write([]byte{byte(function - lastFunction)})
		lastFunction = function
	}

	cw.file.WriteAt([]byte{FUNCTIONS_INDEXES_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeFunctions() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("FUNC")

	cw.writeInstructions(&cw.codeGenerator.FunctionsInstructions)

	cw.file.WriteAt([]byte{FUNCTIONS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeConstantsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("CNST")

	cw.writeStringsSegment()
	cw.writeIntsSegment()
	cw.writeFloatsSegment()

	cw.file.WriteAt([]byte{CONSTANTS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeStringsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("STRS")

	for i := 0; i < len(cw.codeGenerator.stringConstants); i++ {
		cw.file.WriteString(cw.codeGenerator.Constants[i].(string))
		cw.file.Write(STRING_TERMINATOR)
	}

	cw.file.WriteAt([]byte{STRINGS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeIntsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("INTS")

	byteSlice := make([]byte, 8)
	for i := len(cw.codeGenerator.stringConstants); i < len(cw.codeGenerator.stringConstants)+len(cw.codeGenerator.intConstants); i++ {
		binary.BigEndian.PutUint64(byteSlice, uint64(cw.codeGenerator.Constants[i].(int64)))
		cw.file.Write(byteSlice)
	}

	cw.file.WriteAt([]byte{INTS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeFloatsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("FLTS")

	byteSlice := make([]byte, 8)
	for i := len(cw.codeGenerator.stringConstants) + len(cw.codeGenerator.intConstants); i < len(cw.codeGenerator.Constants); i++ {
		binary.BigEndian.PutUint64(byteSlice, math.Float64bits(cw.codeGenerator.Constants[i].(float64)))
		cw.file.Write(byteSlice)
	}

	cw.file.WriteAt([]byte{FLOATS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}
