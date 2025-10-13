package repository

import (
	"sync"

	"github.com/dmnAlex/shortener/internal/model/errx"
)

type URLRepository interface {
	Save(shortID, url string) error
	Find(shortID string) (string, error)
}

type inMemoryRepo struct {
	urls map[string]string
	mu   sync.RWMutex
}

func NewInMemoryRepo() URLRepository {
	return &inMemoryRepo{
		urls: make(map[string]string),
	}
}

func (r *inMemoryRepo) Save(shortID, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.urls[shortID] = url
	return nil
}

func (r *inMemoryRepo) Find(shortID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	url, exists := r.urls[shortID]
	if !exists {
		return "", errx.ErrNotFound
	}
	return url, nil
}
