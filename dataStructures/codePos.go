package dataStructures

import "fmt"

type CodePos struct {
	File *string
	Line uint
	StartChar uint
	EndChar uint
}

func (cp CodePos) String() string {
	return fmt.Sprintf("%s %d:%d", *cp.File, cp.Line, cp.StartChar)
}
