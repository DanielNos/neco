package dataStructures

type Stack struct {
	Top    *StackNode
	Bottom *StackNode
	Size   int
}

type StackNode struct {
	Value    any
	Previous *StackNode
}

func NewStack() *Stack {
	return &Stack{nil, nil, 0}
}

func (s *Stack) Push(value any) {
	s.Top = &StackNode{value, s.Top}
	if s.Bottom == nil {
		s.Bottom = s.Top
	}
	s.Size++
}

func (s *Stack) Pop() any {
	if s.Top == nil {
		return nil
	}

	value := s.Top.Value
	s.Top = s.Top.Previous
	s.Size--

	if s.Top == nil {
		s.Bottom = nil
	}

	return value
}
