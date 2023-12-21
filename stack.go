package main

type Stack struct {
	top *StackNode
	size int
}

type StackNode struct {
	value interface{}
	previous *StackNode
}

func NewStack() *Stack {
	return &Stack{nil, 0}
}

func (s *Stack) Push(value interface{}) {
	s.top = &StackNode{value, s.top}
}

func (s *Stack) Pop() interface{} {
	if s.top == nil {
		return nil
	}

	value := s.top.value
	s.top = s.top.previous

	return value
}
