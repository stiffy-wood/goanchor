package collections

type Queue[T any] struct {
	first  *qNode[T]
	last   *qNode[T]
	length int
}

type qNode[T any] struct {
	prev  *qNode[T]
	next  *qNode[T]
	value T
}

func NewQueue[T any](init ...T) *Queue[T] {
	q := &Queue[T]{}

	for _, i := range init {
		q.Enqueue(i)
	}

	return q
}

func (q *Queue[T]) Enqueue(item T) {
	n := &qNode[T]{
		prev:  q.last,
		value: item,
	}
	q.last = n
	q.length++
}

func (q *Queue[T]) Dequeue() T {
	v := q.first.value
	q.first = q.first.next
	q.length--
	return v
}

func (q *Queue[T]) Peek() T {
	return q.first.value
}

func (q *Queue[T]) GetLength() int {
	return q.length
}
