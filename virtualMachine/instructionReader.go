package virtualMachine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/DanielNos/neco/errors"
	"github.com/DanielNos/neco/logger"
)

type InstructionReader struct {
	filePath  string
	bytes     []byte
	byteIndex int

	virtualMachine *VirtualMachine
}

var NO_ARGS = []int{}

const OFFSET_BYTE_MASK = byte(0b0111_1111)

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
		logger.Fatal(errors.READ_PROGRAM, "Can't "+err.Error()+".")
	}

	// Invalid magic number
	if ir.bytes[0] != 'N' || ir.bytes[1] != 'e' || ir.bytes[2] != 'C' || ir.bytes[3] != 'o' {
		logger.Fatal(errors.READ_PROGRAM, "File isn't a NeCo binary or is corrupted.")
	}

	// Incompatible version
	if ir.bytes[5] > VERSION_MAJOR || ir.bytes[6] > VERSION_MINOR || ir.bytes[7] > VERSION_PATCH {
		logger.Fatal(errors.INCOMPATIBLE_VERSION, fmt.Sprintf("Incompatible version. Binary version is %d.%d.%d, your NeCo version is %d.%d.%d.", ir.bytes[5], ir.bytes[6], ir.bytes[7], VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH))
	}

	ir.byteIndex = 8

	// Read segments
	ir.readConstants()
	ir.readCode()
}

func (ir *InstructionReader) readConstants() {
	ir.byteIndex += 4

	ir.readStringConstants()
	ir.readIntConstants()
	ir.readFloatConstants()
}

func (ir *InstructionReader) readStringConstants() {
	ir.byteIndex++

	// Collect header
	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	segmentEnd := ir.byteIndex + segmentSize

	// Collect strings
	str := []byte{}

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

	// Collect header
	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	segmentEnd := ir.byteIndex + segmentSize

	// Collect ints
	for ir.byteIndex < segmentEnd {
		var integer int64
		binary.Read(bytes.NewReader(ir.bytes[ir.byteIndex:ir.byteIndex+8]), binary.BigEndian, &integer)
		ir.virtualMachine.Constants = append(ir.virtualMachine.Constants, integer)
		ir.byteIndex += 8
	}
}

func (ir *InstructionReader) readFloatConstants() {
	ir.byteIndex++

	// Collect header
	segmentSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	segmentEnd := ir.byteIndex + segmentSize

	// Collect floats
	for ir.byteIndex < segmentEnd {
		floatBits := binary.BigEndian.Uint64(ir.bytes[ir.byteIndex : ir.byteIndex+8])
		ir.virtualMachine.Constants = append(ir.virtualMachine.Constants, math.Float64frombits(floatBits))

		ir.byteIndex += 8
	}
}

func (ir *InstructionReader) readCode() {
	ir.byteIndex += 4

	ir.readMetaData()
	ir.readGlobals()
	ir.readFunctionIndexes()
	ir.readFunctions()
}

func (ir *InstructionReader) readMetaData() {
	ir.byteIndex++

	// Move past header
	ir.byteIndex += 3

	// Collect first line number
	ir.virtualMachine.firstLine = int(ir.bytes[ir.byteIndex])
	ir.byteIndex++
}

func (ir *InstructionReader) readGlobals() {
	ir.byteIndex++

	// Collect header
	codeSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	endIndex := ir.byteIndex + codeSize

	ir.readInstructions(&ir.virtualMachine.GlobalsInstructions, endIndex)
}

func (ir *InstructionReader) readFunctionIndexes() {
	ir.byteIndex++

	// Collect header
	functionsSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	// Allocate slice
	ir.virtualMachine.functions = make([]int, functionsSize)

	lastFunction := 0
	for i := 0; i < functionsSize; i++ {
		ir.virtualMachine.functions[i] = lastFunction + int(ir.bytes[ir.byteIndex])
		lastFunction = ir.virtualMachine.functions[i]
		ir.byteIndex++
	}
}

func (ir *InstructionReader) readFunctions() {
	ir.byteIndex++

	// Collect header
	codeSize := byte3ToInt(ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1], ir.bytes[ir.byteIndex+2])
	ir.byteIndex += 3

	endIndex := ir.byteIndex + codeSize

	ir.readInstructions(&ir.virtualMachine.FunctionsInstructions, endIndex)
}

func (ir *InstructionReader) readInstructions(target *[]ExpandedInstruction, endIndex int) {
	for ir.byteIndex < endIndex {
		instructionType := ir.bytes[ir.byteIndex]

		// 1 argument 2 byte instruction
		if instructionType <= IT_JumpIfTrueEx {
			ir.byteIndex++
			*target = append(*target, ExpandedInstruction{instructionType, []int{int(binary.LittleEndian.Uint16([]byte{ir.bytes[ir.byteIndex], ir.bytes[ir.byteIndex+1]}))}})
			ir.byteIndex++
			// 1 argument 1 byte instruction
		} else if instructionType <= IT_JumpIfTrue {
			ir.byteIndex++

			// Declarator of composite variable
			if IsCompositeDeclarator(instructionType) {
				// Add declare list instruction with id argument
				*target = append(*target, ExpandedInstruction{instructionType, []int{int(ir.bytes[ir.byteIndex])}})
				ir.byteIndex++

				// Add instructions for declaring composite sub-types types without arguments
				for IsCompositeDeclarator(ir.bytes[ir.byteIndex]) {
					*target = append(*target, ExpandedInstruction{ir.bytes[ir.byteIndex], NO_ARGS})
					ir.byteIndex++
				}

				// Add inner-most type instructions without arguments
				*target = append(*target, ExpandedInstruction{ir.bytes[ir.byteIndex], NO_ARGS})

				// Normal instruction
			} else {
				*target = append(*target, ExpandedInstruction{instructionType, []int{int(ir.bytes[ir.byteIndex])}})
			}

			// 0 argument instruction
		} else if instructionType < IT_LineOffset {
			*target = append(*target, ExpandedInstruction{instructionType, NO_ARGS})
			// Line offset
		} else {
			*target = append(*target, ExpandedInstruction{IT_LineOffset, []int{int((ir.bytes[ir.byteIndex] & OFFSET_BYTE_MASK)) + 1}})
		}

		ir.byteIndex++
	}
}
