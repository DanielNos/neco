package virtualMachine

type Stack struct {
	items    []any
	size     int
	capacity int
}

func NewStack(size int) *Stack {
	return &Stack{make([]any, size), 0, size}
}

func (s *Stack) Push(value any) {
	if s.size == s.capacity {
		panic("STACK OVERFLOW!")
	}

	s.items[s.size] = value
	s.size++
}

func (s *Stack) Pop() any {
	s.size--
	return s.items[s.size]
}

func (s *Stack) Top() *any {
	return &s.items[s.size-1]
}

func (s *Stack) Previous() *any {
	return &s.items[s.size-2]
}
