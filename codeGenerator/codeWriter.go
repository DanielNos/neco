package codeGenerator

import (
	"encoding/binary"
	"math"
	"os"
)

const CONSTANTS_SEGMENT = 0
const STRINGS_SEGMENT = 0
const INTS_SEGMENT = 1
const FLOATS_SEGMENT = 2

const CODE_SEGMENT = 1

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

	for _, instruction := range cw.codeGenerator.instructions {
		cw.file.Write([]byte{instruction.InstructionType})
		cw.file.Write(instruction.InstructionValue)
	}

	cw.file.WriteAt([]byte{CODE_SEGMENT}, startPos)
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
		cw.file.WriteString(cw.codeGenerator.constants[i].(string))
		cw.file.Write([]byte{0})
	}

	cw.file.WriteAt([]byte{STRINGS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeIntsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("INTS")

	byteSlice := make([]byte, 8)
	for i := len(cw.codeGenerator.stringConstants); i < len(cw.codeGenerator.stringConstants)+len(cw.codeGenerator.intConstants); i++ {
		binary.BigEndian.PutUint64(byteSlice, uint64(cw.codeGenerator.constants[i].(int64)))
		cw.file.Write(byteSlice)
	}

	cw.file.WriteAt([]byte{INTS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}

func (cw *CodeWriter) writeFloatsSegment() {
	startPos := cw.getFilePosition()
	cw.file.WriteString("FLTS")

	byteSlice := make([]byte, 8)
	for i := len(cw.codeGenerator.stringConstants) + len(cw.codeGenerator.intConstants); i < len(cw.codeGenerator.constants); i++ {
		binary.BigEndian.PutUint64(byteSlice, math.Float64bits(cw.codeGenerator.constants[i].(float64)))
		cw.file.Write(byteSlice)
	}

	cw.file.WriteAt([]byte{FLOATS_SEGMENT}, startPos)
	cw.file.WriteAt(int64ToByte3(cw.getFilePosition()-startPos-4), startPos+1)
}
