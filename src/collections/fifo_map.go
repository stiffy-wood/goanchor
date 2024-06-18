package collections

type FifoMap[K comparable, V any] struct {
	first *fmNode[K, V]
	last  *fmNode[K, V]
	fMap  map[K]*fmNode[K, V]
}

type fmNode[K comparable, V any] struct {
	prev  *fmNode[K, V]
	next  *fmNode[K, V]
	value V
	key   K
}

func NewFifoMap[K comparable, V any]() *FifoMap[K, V] {
	return &FifoMap[K, V]{}
}

func (m *FifoMap[K, V]) Enqueue(key K, value V) {
	n := &fmNode[K, V]{
		prev:  m.last,
		value: value,
		key:   key,
	}
	m.last = n
	m.fMap[key] = n
}

func (m *FifoMap[K, V]) Dequeue() V {
	v := m.first.value
	delete(m.fMap, m.first.key)
	m.first = m.first.next
	return v
}

func (m *FifoMap[K, V]) Peek() V {
	return m.first.value
}

func (m *FifoMap[K, V]) Get(key K) V {
    return m.fMap[key].value
}

func (m *FifoMap[K, V]) Exists(key K) bool {
    _, exists := m.fMap[key]
    return exists
}

func (m *FifoMap[K, V]) GetLength() int {
	return len(m.fMap)
}
