package codeOptimizer

import (
	VM "github.com/DanielNos/neco/virtualMachine"
)

const IGNORE_INSTRUCTION byte = 255

func calculateReductionBack(instructions *[]VM.Instruction, instructionIndex int) int {
	reduction := 0

	for i := instructionIndex; i > instructionIndex-int((*instructions)[instructionIndex].InstructionValue[0]); i-- {
		if (*instructions)[i].InstructionType == IGNORE_INSTRUCTION {
			reduction++
		}
	}

	return reduction
}

func calculateReductionForward(instructions *[]VM.Instruction, instructionIdnex int) int {
	reduction := 0

	for i := instructionIdnex; i < instructionIdnex+int((*instructions)[instructionIdnex].InstructionValue[0])+1; i++ {
		if (*instructions)[i].InstructionType == IGNORE_INSTRUCTION {
			reduction++
		}
	}

	return reduction
}

func Optimize(instructions *[]VM.Instruction, functions []int) {
	// Append instruction buffer to the end
	for i := 0; i < 2; i++ {
		*instructions = append(*instructions, VM.Instruction{255, []byte{}})
	}

	// Optimize instructions
	for i := 0; i < len(*instructions); i++ {
		// Combine line offsets
		if (*instructions)[i].InstructionType == VM.IT_LineOffset && (*instructions)[i+1].InstructionType == VM.IT_LineOffset {
			(*instructions)[i+1].InstructionValue[0] += (*instructions)[i].InstructionValue[0]
			(*instructions)[i].InstructionType = IGNORE_INSTRUCTION
			continue
		}

		// Zero distance jump
		if (*instructions)[i].InstructionType == VM.IT_Jump && (*instructions)[i].InstructionValue[0] == 0 {
			(*instructions)[i].InstructionType = IGNORE_INSTRUCTION
			continue
		}
	}

	// Calculate all removed (*instructions) between jumps and their destinations and reduce jump by that amount
	for i := 0; i < len((*instructions)); i++ {
		if VM.IsJumpForward((*instructions)[i].InstructionType) {
			(*instructions)[i].InstructionValue[0] -= byte(calculateReductionForward(instructions, i))
		} else if (*instructions)[i].InstructionType == VM.IT_JumpBack {
			(*instructions)[i].InstructionValue[0] -= byte(calculateReductionBack(instructions, i))
		}
	}

	// Adjust function positions
	for functionIndex, functionPosition := range functions {
		toReduce := 0
		for i := functionPosition; i >= 0; i-- {
			if (*instructions)[i].InstructionType == IGNORE_INSTRUCTION {
				toReduce++
			}
		}

		functions[functionIndex] -= toReduce
	}
}
