package virtualMachine

type Stack struct {
	items    []interface{}
	size     int
	capacity int
}

func NewStack(size int) *Stack {
	return &Stack{make([]interface{}, size), 0, size}
}

func (s *Stack) Push(value interface{}) {
	if s.size == s.capacity {
		panic("STACK OVERFLOW!")
	}

	s.items[s.size] = value
	s.size++
}

func (s *Stack) Pop() interface{} {
	s.size--
	return s.items[s.size]
}

func (s *Stack) Top() *interface{} {
	return &s.items[s.size-1]
}

func (s *Stack) Previous() *interface{} {
	return &s.items[s.size-2]
}
