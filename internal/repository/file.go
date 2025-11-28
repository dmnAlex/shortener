package repository

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/utils"
)

type URLRepository interface {
	Save(userID, url string) (string, error)
	SaveBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)
	Find(shortID string) (string, error)
	FindAll(userID string) ([]model.UserURLsResponse, error)
	Ping() error
	Close() error
}

type fileRepo struct {
	mu              sync.RWMutex
	path            string
	ShortToOriginal map[string]string
	OriginalToShort map[string]string
	UserToShorts    map[string]map[string]string // userID -> shortID -> originalURL
}

func NewFileRepo(path string) (*fileRepo, error) {
	r := &fileRepo{
		path:            path,
		ShortToOriginal: make(map[string]string),
		OriginalToShort: make(map[string]string),
		UserToShorts:    make(map[string]map[string]string),
	}

	if err := r.loadFromFile(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *fileRepo) Save(userID, originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShort, exists := r.OriginalToShort[originalURL]; exists {
		if _, ok := r.UserToShorts[userID]; !ok {
			r.UserToShorts[userID] = make(map[string]string)
		}
		if _, has := r.UserToShorts[userID][existingShort]; !has {
			r.UserToShorts[userID][existingShort] = originalURL
		}

		if err := r.saveToFile(); err != nil {
			return "", err
		}

		return existingShort, errx.ErrConflict
	}

	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	r.ShortToOriginal[shortID] = originalURL
	r.OriginalToShort[originalURL] = shortID

	if _, ok := r.UserToShorts[userID]; !ok {
		r.UserToShorts[userID] = make(map[string]string)
	}
	r.UserToShorts[userID][shortID] = originalURL

	if err := r.saveToFile(); err != nil {
		return "", err
	}

	return shortID, nil
}

func (r *fileRepo) SaveBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	var res []model.ShortenBatchResponse
	for _, item := range batch {
		shortID, err := r.Save(userID, item.OriginalURL)
		if err != nil && !errors.Is(err, errx.ErrConflict) {
			return nil, err
		}

		res = append(res, model.ShortenBatchResponse{CorrelationID: item.CorrelationID, ShortURL: shortID})
	}

	return res, nil
}

func (r *fileRepo) Find(shortID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.ShortToOriginal[shortID], nil
}

func (r *fileRepo) FindAll(userID string) ([]model.UserURLsResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := []model.UserURLsResponse{}
	if shorts, ok := r.UserToShorts[userID]; ok {
		for shortID, originalURL := range shorts {
			res = append(res, model.UserURLsResponse{ShortURL: shortID, OriginalURL: originalURL})
		}
	}

	return res, nil
}

func (r *fileRepo) Ping() error {
	return nil
}

func (r *fileRepo) Close() error {
	return nil
}

func (r *fileRepo) saveToFile() error {
	if r.path == "" {
		return nil
	}

	file, err := os.Create(r.path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(r)
}

func (r *fileRepo) loadFromFile() error {
	if r.path == "" {
		return nil
	}

	file, err := os.Open(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(r)
}
