package ringbuf

type Ring[T any] struct {
	Data []T
	Head int
}

func New[T any](size int) *Ring[T] {
	return &Ring[T]{
		Data: make([]T, size),
		Head: 0,
	}
}

func (r *Ring[T]) PushFront(v T) *Ring[T] {
	r.Head = r.Head - 1
	if r.Head < 0 {
		r.Head = len(r.Data) - 1
	}
	r.Data[r.Head] = v
	return r
}

func (r *Ring[T]) WalkFirstN(count int, fn func(T)) {
	for i := 0; i < count; i++ {
		fn(r.Data[(r.Head+i)%len(r.Data)])
	}
}

func (r *Ring[T]) GetN(i int) T {
	idx := r.Head + i
	if idx < 0 {
		idx = len(r.Data) + idx
	}
	return r.Data[(idx)%len(r.Data)]
}

func (r *Ring[T]) SetN(i int, val T) {
	idx := r.Head + i
	if idx < 0 {
		idx = len(r.Data) + idx
	}
	r.Data[(idx)%len(r.Data)] = val
}

func (r *Ring[T]) Len() int {
	return len(r.Data)
}
