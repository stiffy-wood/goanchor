package collections

import (
	"sync"
)

type Set[T comparable] struct {
	lock sync.RWMutex
	hash map[T]struct{}
}

func NewSet[T comparable](init ...T) *Set[T] {
	s := &Set[T]{
		lock: sync.RWMutex{},
		hash: make(map[T]struct{}),
	}

	for _, v := range init {
		s.Insert(v)
	}

	return s
}

func (s *Set[T]) Insert(item T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.hash[item] = struct{}{}
}

func (s *Set[T]) Has(item T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	_, exists := s.hash[item]
	return exists
}

func (s *Set[T]) Length() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.hash)
}

func (s *Set[T]) Remove(item T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.hash, item)
}

func (s *Set[T]) InserIfNotExist(item T) (exists bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exists = s.hash[item]
	if exists {
		return true
	}
	s.hash[item] = struct{}{}
	return false
}
