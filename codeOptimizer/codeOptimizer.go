package codeOptimizer

import (
	VM "neco/virtualMachine"
)

func Optimize(instructions []VM.Instruction) {
	for i := 0; i < len(instructions)-1; i++ {
		// Combine line offsets
		if instructions[i].InstructionType == VM.IT_LineOffset && instructions[i+1].InstructionType == VM.IT_LineOffset {
			instructions[i+1].InstructionValue[0] += instructions[i].InstructionValue[0]
			instructions[i].InstructionType = 255
			continue
		}

		// Zero distance jump
		if instructions[i].InstructionType == VM.IT_Jump && instructions[i].InstructionValue[0] == 0 {
			instructions[i].InstructionType = 255
			continue
		}		
	}
}
