package dataStructures

import "fmt"

type CodePos struct {
	File      *string
	StartLine uint
	EndLine   uint
	StartChar uint
	EndChar   uint
}

// Combines two CodePositions.
func (cp CodePos) Combine(codePos *CodePos) *CodePos {
	if cp.File != codePos.File {
		panic("Combine used on code positions with different files.")
	}

	newCodePos := &CodePos{cp.File, min(cp.StartLine, codePos.StartLine), max(cp.EndLine, codePos.EndLine), 0, 0}

	if cp.StartLine != codePos.StartLine || cp.EndLine != codePos.EndLine {
		if cp.StartLine > codePos.StartLine || cp.EndLine > codePos.EndLine {
			newCodePos.EndChar = cp.EndChar
			newCodePos.StartChar = codePos.StartChar
		} else {
			newCodePos.EndChar = codePos.EndChar
			newCodePos.StartChar = cp.StartChar
		}
	} else {
		newCodePos.StartChar = min(cp.StartChar, codePos.StartChar)
		newCodePos.EndChar = max(cp.EndChar, codePos.EndChar)
	}

	return newCodePos
}

func (cp CodePos) String() string {
	return fmt.Sprintf("%s %d:%d %d:%d", *cp.File, cp.StartLine, cp.StartChar, cp.EndLine, cp.EndChar)
}
