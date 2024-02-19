package codeOptimizer

import (
	VM "neco/virtualMachine"
)

const IGNORE_INSTRUCTION byte = 255

func Optimize(instructions []VM.Instruction) {
	for i := 0; i < len(instructions)-1; i++ {
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
		if instructions[i].InstructionType == VM.IT_LoadConstRegA && instructions[i+1].InstructionType == VM.IT_PushRegAToArgStack {
			instructions[i].InstructionType = VM.IT_LoadConstArgStack
			instructions[i+1].InstructionType = IGNORE_INSTRUCTION
			i++
			continue
		}

		// Load variable directly do argument stack
		if instructions[i].InstructionType == VM.IT_LoadRegA && instructions[i+1].InstructionType == VM.IT_PushRegAToArgStack {
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

		if instructions[i].InstructionType == VM.IT_LoadRegA && instructions[i+1].InstructionType == VM.IT_CopyRegAToE && instructions[i+2].InstructionType == VM.IT_LoadRegA && instructions[i+3].InstructionType == VM.IT_CopyRegEToB {
			instructions[i].InstructionType = VM.IT_LoadRegB
			instructions[i+1].InstructionType = IGNORE_INSTRUCTION
			instructions[i+3].InstructionType = IGNORE_INSTRUCTION
			i += 3
			continue
		}
	}
}
