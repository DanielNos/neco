package codegenerator

import (
	"encoding/binary"
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

func Int32ToByte3(value int) []byte {
	intBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(intBytes, uint32(value))
	return intBytes[1:4]
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

func (cw *CodeWriter) writeSegment(segmentCode byte, segmentData *[]byte) {
	cw.file.Write([]byte{segmentCode})

	cw.file.Write(Int32ToByte3(len(*segmentData)))
	cw.file.Write(*segmentData)
}

func (cw *CodeWriter) writeConstantsSegment() {
	cw.file.WriteString("CNST")

	cw.writeStringsSegment()
}

func (cw *CodeWriter) writeCodeSegment() {
	cw.file.WriteString("CODE")

	for _, instruction := range cw.codeGenerator.instructions {
		cw.file.Write([]byte{instruction.InstructionType, instruction.ValueA.(byte), instruction.ValueB.(byte), instruction.ValueC.(byte)})
	}
}

func (cw *CodeWriter) writeStringsSegment() {
	stringsSegment := []byte{}

	for _, index := range cw.codeGenerator.stringConstants {
		stringsSegment = append(stringsSegment, []byte(cw.codeGenerator.Constants[index].Value.(string))...)
		stringsSegment = append(stringsSegment, 0)
	}

	cw.writeSegment(0, &stringsSegment)
}
