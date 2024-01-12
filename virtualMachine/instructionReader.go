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

	virtualMachine *VirtualMachine
}

func NewInstructionReader(filePath string, virtualMachine *VirtualMachine) *InstructionReader {
	return &InstructionReader{filePath, nil, 0, virtualMachine}
}

func byte3ToInt(byte1, byte2, byte3 byte) int {
	paddedBytes := make([]byte, 8)

	paddedBytes[5] = byte1
	paddedBytes[6] = byte2
	paddedBytes[7] = byte3

	return int(binary.BigEndian.Uint64(paddedBytes))
}

func (ir *InstructionReader) Read() {
	// Read file
	var err error
	ir.bytes, err = os.ReadFile(ir.filePath)

	// Couldn't read file
	if err != nil {
		logger.Fatal(errors.ERROR_READ_PROGRAM, "Can't read file.")
	}

	// Invalid magic number
	if ir.bytes[0] != 'N' || ir.bytes[1] != 'E' || ir.bytes[2] != 'C' || ir.bytes[3] != 'O' {
		logger.Fatal(errors.ERROR_READ_PROGRAM, "File isn't a NeCo binary.")
	}

	// Incompatible version
	if ir.bytes[5] > VERSION_MAJOR || ir.bytes[6] > VERSION_MINOR || ir.bytes[7] > VERSION_PATCH {
		logger.Fatal(errors.ERROR_INCOMPATIBLE_VERSION, fmt.Sprintf("Incompatible version. Binary version is %d.%d.%d, your NeCo version is %d.%d.%d.", ir.bytes[5], ir.bytes[6], ir.bytes[7], VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH))
	}

	ir.byteIndex = 8

	// Read segments
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
			ir.virtualMachine.Constants = append(ir.virtualMachine.Constants, string(str))
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

	segmentEnd := ir.byteIndex + segmentSize

	for ir.byteIndex < segmentEnd {
		var integer int64
		binary.Read(bytes.NewReader(ir.bytes[ir.byteIndex:ir.byteIndex+8]), binary.BigEndian, &integer)
		ir.virtualMachine.Constants = append(ir.virtualMachine.Constants, integer)
		ir.byteIndex += 8
	}
}

func (ir *InstructionReader) readFloatConstants() {
	ir.byteIndex++

	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	segmentEnd := ir.byteIndex + segmentSize

	for ir.byteIndex < segmentEnd {
		floatBits := binary.BigEndian.Uint64(ir.bytes[ir.byteIndex : ir.byteIndex+8])
		ir.virtualMachine.Constants = append(ir.virtualMachine.Constants, math.Float64frombits(floatBits))

		ir.byteIndex += 8
	}
}

func (ir *InstructionReader) readInstructions() {
	ir.byteIndex++

	codeSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	endIndex := ir.byteIndex + codeSize

	// Collect first line number
	ir.virtualMachine.Line = uint(ir.bytes[ir.byteIndex]) - 128
	ir.byteIndex++

	offsetByteMask := byte(0b0111_1111)

	for ir.byteIndex < endIndex {
		switch ir.bytes[ir.byteIndex] {
		case IT_LoadConstant:
			ir.byteIndex++
			ir.virtualMachine.Instructions = append(ir.virtualMachine.Instructions, Instruction{IT_LoadConstant, []byte{ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1]}})
			ir.byteIndex++
		case IT_CallBuiltInFunction:
			ir.byteIndex++
			ir.virtualMachine.Instructions = append(ir.virtualMachine.Instructions, Instruction{IT_CallBuiltInFunction, []byte{ir.bytes[ir.byteIndex]}})
		case IT_Halt:
			ir.byteIndex++
			ir.virtualMachine.Instructions = append(ir.virtualMachine.Instructions, Instruction{IT_Halt, []byte{ir.bytes[ir.byteIndex]}})
		case IT_Push:
			ir.byteIndex++
			ir.virtualMachine.Instructions = append(ir.virtualMachine.Instructions, Instruction{IT_Push, []byte{ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1]}})
			ir.byteIndex++

		default:
			ir.virtualMachine.Instructions = append(ir.virtualMachine.Instructions, Instruction{IT_LineOffset, []byte{ir.bytes[ir.byteIndex] & offsetByteMask}})
		}

		ir.byteIndex++
	}
}
