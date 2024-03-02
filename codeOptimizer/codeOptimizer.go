package codeOptimizer

import (
	"fmt"
	VM "neco/virtualMachine"
)

const IGNORE_INSTRUCTION byte = 255

func Optimize(instructions *[]VM.Instruction) {
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
			(*instructions)[i].InstructionType = 255
			fmt.Printf("DELETED JUMP %d\n ", i)
			continue
		}

		// Load const directly to list
		if (*instructions)[i].InstructionType == VM.IT_LoadConst && (*instructions)[i+1].InstructionType == VM.IT_AppendToList {
			(*instructions)[i].InstructionType = VM.IT_LoadConstToList
			(*instructions)[i+1].InstructionType = 255
			continue
		}
	}

	// Adjust jumps for removed (*instructions)
	for i := 0; i < len((*instructions)); i++ {
		// Jump forward
		if VM.IsJumpForward((*instructions)[i].InstructionType) {
			// Calculate all removed (*instructions) between jump and it's destination
			reduction := 0

			for j := i; j < i+int((*instructions)[i].InstructionValue[0])+1; j++ {
				if (*instructions)[j].InstructionType == IGNORE_INSTRUCTION {
					reduction++
				}
			}

			// Reduce jump by that amount
			(*instructions)[i].InstructionValue[0] -= byte(reduction)

			// Jump back
		} else if (*instructions)[i].InstructionType == VM.IT_JumpBack {
			// Calculate all removed (*instructions) between jump and it's destination

			reduction := 0

			for j := i; j > i-int((*instructions)[i].InstructionValue[0]); j-- {
				if (*instructions)[j].InstructionType == IGNORE_INSTRUCTION {
					reduction++
				}
			}

			// Reduce jump by that amount
			(*instructions)[i].InstructionValue[0] -= byte(reduction)
		}
	}
}
