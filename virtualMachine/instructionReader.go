package virtualMachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"neko/errors"
	"neko/logger"
	"os"
)

type InstructionReader struct {
	filePath  string
	bytes     []byte
	byteIndex int

	instructions *[]Instruction
	constants    *[]interface{}
}

func NewInstructionReader(filePath string, instructions *[]Instruction, constants *[]interface{}) *InstructionReader {
	return &InstructionReader{filePath, nil, 0, instructions, constants}
}

func byte3ToInt(byte1, byte2, byte3 byte) int {
	paddedBytes := make([]byte, 8)

	paddedBytes[5] = byte1
	paddedBytes[6] = byte2
	paddedBytes[7] = byte3

	return int(binary.BigEndian.Uint64(paddedBytes))
}

func (ir *InstructionReader) Read() {
	var err error
	ir.bytes, err = os.ReadFile(ir.filePath)

	if err != nil {
		logger.Fatal(errors.ERROR_READ_PROGRAM, "Can't read file.")
	}

	if ir.bytes[0] != 'N' || ir.bytes[1] != 'E' || ir.bytes[2] != 'C' || ir.bytes[3] != 'O' {
		logger.Fatal(errors.ERROR_READ_PROGRAM, "File isn't a NeCo binary.")
	}

	if ir.bytes[5] > VERSION_MAJOR || ir.bytes[6] > VERSION_MINOR || ir.bytes[7] > VERSION_PATCH {
		logger.Fatal(errors.ERROR_INCOMPATIBLE_VERSION, fmt.Sprintf("Incompatible version. Binary version is %d.%d.%d, your NeCo version is %d.%d.%d.", ir.bytes[5], ir.bytes[6], ir.bytes[7], VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH))
	}

	ir.byteIndex = 8

	ir.readConstants()
	ir.readInstructions()
}

func (ir *InstructionReader) readConstants() {
	ir.byteIndex += 4

	ir.readStringConstants()
	ir.readIntConstants()
	ir.readFloatConstants()
}

func (ir *InstructionReader) readStringConstants() {
	ir.byteIndex++

	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	str := []byte{}

	segmentEnd := ir.byteIndex + segmentSize

	for ir.byteIndex < segmentEnd {
		if ir.bytes[ir.byteIndex] == 0 {
			*ir.constants = append(*ir.constants, string(str))
			str = []byte{}
		} else {
			str = append(str, ir.bytes[ir.byteIndex])
		}
		ir.byteIndex++
	}
}

func (ir *InstructionReader) readIntConstants() {
	ir.byteIndex++

	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	integerBytes := []byte{}

	segmentEnd := ir.byteIndex + segmentSize

	for ir.byteIndex < segmentEnd {
		if ir.bytes[ir.byteIndex] == 0 {
			var integer int64
			binary.Read(bytes.NewReader(integerBytes), binary.BigEndian, &integer)
			*ir.constants = append(*ir.constants, integer)
			integerBytes = []byte{}
		} else {
			integerBytes = append(integerBytes, ir.bytes[ir.byteIndex])
		}
		ir.byteIndex++
	}
}

func (ir *InstructionReader) readFloatConstants() {
	ir.byteIndex++

	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	integerBytes := []byte{}

	segmentEnd := ir.byteIndex + segmentSize

	for ir.byteIndex < segmentEnd {
		if ir.bytes[ir.byteIndex] == 0 {
			floatBits := binary.BigEndian.Uint64(integerBytes)
			*ir.constants = append(*ir.constants, math.Float64frombits(floatBits))
			integerBytes = []byte{}
		} else {
			integerBytes = append(integerBytes, ir.bytes[ir.byteIndex])
		}
		ir.byteIndex++
	}
}

func (ir *InstructionReader) readInstructions() {
	ir.byteIndex++

	codeSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	endIndex := ir.byteIndex + codeSize

	for ir.byteIndex < endIndex {
		switch ir.bytes[ir.byteIndex] {
		case IT_LoadConstant:
			ir.byteIndex++
			*ir.instructions = append(*ir.instructions, Instruction{IT_LoadConstant, []byte{ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1]}})
			ir.byteIndex++
		case IT_CallBuiltInFunction:
			ir.byteIndex++
			*ir.instructions = append(*ir.instructions, Instruction{IT_CallBuiltInFunction, []byte{ir.bytes[ir.byteIndex]}})
		case IT_Halt:
			ir.byteIndex++
			*ir.instructions = append(*ir.instructions, Instruction{IT_Halt, []byte{ir.bytes[ir.byteIndex]}})
		}

		ir.byteIndex++
	}
}
