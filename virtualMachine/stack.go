package virtualMachine

type Stack[T any] struct {
	items    []T
	size     int
	capacity int
}

func NewStack[T any](size int) *Stack[T] {
	return &Stack[T]{make([]T, size), 0, size}
}

func (s *Stack[T]) Push(value T) {
	if s.size == s.capacity {
		panic("Stack overflow.")
	}

	s.items[s.size] = value
	s.size++
}

func (s *Stack[T]) Pop() T {
	s.size--
	return s.items[s.size]
}

func (s *Stack[T]) Top() *T {
	return &s.items[s.size-1]
}

func (s *Stack[T]) Previous() *T {
	return &s.items[s.size-2]
}
