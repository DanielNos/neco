package dataStructures

import "fmt"

type CodePos struct {
	File      *string
	StartLine uint
	EndLine   uint
	StartChar uint
	EndChar   uint
}

func (cp CodePos) String() string {
	return fmt.Sprintf("%s %d:%d %d:%d", *cp.File, cp.StartLine, cp.StartChar, cp.EndLine, cp.EndChar)
}
