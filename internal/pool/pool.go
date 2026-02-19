package pool

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	items []T
	newFn func() T
}

func New[T Resettable](factory func() T) *Pool[T] {
	return &Pool[T]{
		newFn: factory,
	}
}

func (p *Pool[T]) Get() T {
	if n := len(p.items); n > 0 {
		item := p.items[n-1]
		p.items = p.items[:n-1]
		return item
	}
	return p.newFn()
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.items = append(p.items, item)
}
