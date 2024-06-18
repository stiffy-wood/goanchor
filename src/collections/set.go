package collections

type Set[T comparable] struct{
    hash map[T]struct{}
}

func NewSet [T comparable](init ...T) *Set[T] {
    s := &Set[T]{
        make(map[T]struct{}),
    }

    for _, v := range init {
        s.Insert(v)
    }

    return s
}

func (s *Set[T]) Insert (item T){
    s.hash[item] = struct{}{}
}

func (s *Set[T]) Has (item T) bool {
    _, exists := s.hash[item]
    return exists
}

func (s *Set[T]) Length() int {
    return len(s.hash)
}

func (s *Set[T]) Remove(item T) {
    delete(s.hash, item)
}
