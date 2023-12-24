package dataStructures

type Stack struct {
	Top *StackNode
	Size int
}

type StackNode struct {
	Value interface{}
	previous *StackNode
}

func NewStack() *Stack {
	return &Stack{nil, 0}
}

func (s *Stack) Push(value interface{}) {
	s.Top = &StackNode{value, s.Top}
}

func (s *Stack) Pop() interface{} {
	if s.Top == nil {
		return nil
	}

	value := s.Top.Value
	s.Top = s.Top.previous

	return value
}
