package stack

//lo vamos a usar para operands, operators y jumps

type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() T {
	if len(s.items) == 0 {
		panic("stack vacio")
	}

	n := len(s.items)

	item := s.items[n-1]
	s.items = s.items[:n-1]

	return item
}

func (s *Stack[T]) Top() T {
	if len(s.items) == 0 {
		panic("stack vacio")
	}
	return s.items[len(s.items)-1]
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}
