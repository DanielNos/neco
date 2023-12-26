package dataStructures

type Stack struct {
	Top *StackNode
	Bottom *StackNode
	Size int
}

type StackNode struct {
	Value interface{}
	Previous *StackNode
}

func NewStack() *Stack {
	return &Stack{nil, nil, 0}
}

func (s *Stack) Push(value interface{}) {
	s.Top = &StackNode{value, s.Top}
	if s.Bottom == nil {
		s.Bottom = s.Top
	}
}

func (s *Stack) Pop() interface{} {
	if s.Top == nil {
		return nil
	}

	value := s.Top.Value
	s.Top = s.Top.Previous

	if s.Top == nil {
		s.Bottom = nil
	}

	return value
}
