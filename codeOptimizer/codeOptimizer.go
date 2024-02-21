package codeOptimizer

import (
	VM "neco/virtualMachine"
)

const IGNORE_INSTRUCTION byte = 255

func Optimize(instructions []VM.Instruction) {
	// Append instruction buffer to the end
	for i := 0; i < 4; i++ {
		instructions = append(instructions, VM.Instruction{255, []byte{}})
	}

	// Optimize instructions
	for i := 0; i < len(instructions); i++ {
		// Combine line offsets
		if instructions[i].InstructionType == VM.IT_LineOffset && instructions[i+1].InstructionType == VM.IT_LineOffset {
			instructions[i+1].InstructionValue[0] += instructions[i].InstructionValue[0]
			instructions[i].InstructionType = IGNORE_INSTRUCTION
			continue
		}

		// Zero distance jump
		if instructions[i].InstructionType == VM.IT_Jump && instructions[i].InstructionValue[0] == 0 {
			instructions[i].InstructionType = IGNORE_INSTRUCTION
			continue
		}

		// Load constant directly do argument stack
		if instructions[i].InstructionType == VM.IT_LoadConstRegA && instructions[i+1].InstructionType == VM.IT_PushOpAToArg {
			instructions[i].InstructionType = VM.IT_LoadConstArgStack
			instructions[i+1].InstructionType = IGNORE_INSTRUCTION
			i++
			continue
		}

		// Load variable directly do argument stack
		if instructions[i].InstructionType == VM.IT_LoadRegA && instructions[i+1].InstructionType == VM.IT_PushOpAToArg {
			instructions[i].InstructionType = VM.IT_LoadArgStack
			instructions[i+1].InstructionType = IGNORE_INSTRUCTION
			i++
			continue
		}

		// Optimize expression safe swapping

		// 0 LOAD_REG_A        1   -->  0 LOAD_REG_B 1
		// 1 COPY_REG_A_TO_E       -->
		// 2 LOAD_REG_A        2   -->  2 LOAD_REG_A 2
		// 3 COPY_REG_E_TO_B       -->

		if instructions[i].InstructionType == VM.IT_LoadRegA && instructions[i+1].InstructionType == VM.IT_CopyOpAToListA && instructions[i+2].InstructionType == VM.IT_LoadRegA && instructions[i+3].InstructionType == VM.IT_CopyListAToOpB {
			instructions[i].InstructionType = VM.IT_LoadRegB
			instructions[i+1].InstructionType = IGNORE_INSTRUCTION
			instructions[i+3].InstructionType = IGNORE_INSTRUCTION
			i += 3
			continue
		}
	}

	// Adjust jumps for removed instructions
	for i := 0; i < len(instructions); i++ {
		// Locate jump
		if VM.IsJumpForward(instructions[i].InstructionType) {
			// Calculate all removed instructions between jump and it's destination
			reduction := 0

			for j := i; j < i+int(instructions[i].InstructionValue[0]); j++ {
				if instructions[j].InstructionType == IGNORE_INSTRUCTION {
					reduction++
				}
			}

			// Reduce jump by that amount
			instructions[i].InstructionValue[0] -= byte(reduction)
		}
	}
}
