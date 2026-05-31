package utils

func NewCircleQueue[T comparable](size int) *CircleQueue[T] {
	return &CircleQueue[T]{
		data: make([]T, size),
		idx:  0,
		size: size,
	}
}

type CircleQueue[T comparable] struct {
	data []T
	idx  int
	size int
}

func (q *CircleQueue[T]) Add(value T) T {
	// only will be a value when there idx goes already went full circle
	// so values a getting removed
	currentOffset := q.idx % q.size
	output := q.data[currentOffset]
	q.data[currentOffset] = value
	return output
}
