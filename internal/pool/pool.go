package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	mu    sync.Mutex
	items []T
	newFn func() T
}

func New[T Resettable](factory func() T) *Pool[T] {
	return &Pool[T]{
		newFn: factory,
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if n := len(p.items); n > 0 {
		item := p.items[n-1]
		p.items = p.items[:n-1]
		return item
	}
	return p.newFn()
}

func (p *Pool[T]) Put(item T) {
	item.Reset()
	p.mu.Lock()
	p.items = append(p.items, item)
	p.mu.Unlock()
}
