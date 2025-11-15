package repository

import (
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
)

type URLRepository interface {
	Save(shortID, url string) error
	Find(shortID string) (string, error)
	Ping() error
}

type fileRepo struct {
	path     string
	mu       sync.RWMutex
	urls     map[string]string
	records  []model.URLRecord
	nextUUID int
}

func NewFileRepo(path string) (URLRepository, error) {
	r := &fileRepo{
		path:     path,
		urls:     make(map[string]string),
		records:  make([]model.URLRecord, 0),
		nextUUID: 1,
	}

	if r.path == "" {
		return r, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return r, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &r.records); err != nil {
		return nil, err
	}

	for _, record := range r.records {
		r.urls[record.ShortURL] = record.OriginalURL
		u, err := strconv.Atoi(record.UUID)
		if err == nil && u >= r.nextUUID {
			r.nextUUID = u + 1
		}
	}

	return r, nil
}

func (r *fileRepo) saveToFile() error {
	if r.path == "" {
		return nil
	}

	data, err := json.Marshal(r.records)
	if err != nil {
		return err
	}

	return os.WriteFile(r.path, data, 0666)
}

func (r *fileRepo) Save(shortID, url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[shortID] = url

	exists := false
	for i := range r.records {
		if r.records[i].ShortURL == shortID {
			r.records[i].OriginalURL = url
			exists = true
			break
		}
	}

	if !exists {
		r.records = append(r.records, model.URLRecord{
			UUID:        strconv.Itoa(r.nextUUID),
			ShortURL:    shortID,
			OriginalURL: url,
		})
		r.nextUUID++
	}

	return r.saveToFile()
}

func (r *fileRepo) Find(shortID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.urls[shortID]
	if !exists {
		return "", errx.ErrNotFound
	}

	return url, nil
}

func (r *fileRepo) Ping() error {
	return nil
}
