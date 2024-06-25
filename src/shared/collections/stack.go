package collections

type Stack[T any] struct {
	top    *sNode[T]
	length int
}

type sNode[T any] struct {
	value T
	prev  *sNode[T]
}

func NewStack[T any](init ...T) *Stack[T] {
	s := &Stack[T]{}
	for _, v := range init {
		s.Push(v)
	}
	return s
}

func (s *Stack[T]) Length() int {
	return s.length
}

func (s *Stack[T]) Peek() T {
	var out T
	if s.top != nil {
		out = s.top.value
	}
	return out
}

func (s *Stack[T]) Pop() T {
	var out T
	if s.top != nil {
		out = s.top.value
	}
	s.top = s.top.prev
	s.length--
	return out
}

func (s *Stack[T]) Push(val T) {
	n := &sNode[T]{
		value: val,
		prev:  s.top,
	}
	s.top = n
	s.length++
}
