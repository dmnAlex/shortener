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
	DeleteURLs(userID string, shortIDs []string) error
}

type fileRepo struct {
	mu             sync.RWMutex
	path           string
	URLRecords     map[string]*model.URLRecord    // shortID -> shortID -> *model.URLRecord
	IdxOriginalURL map[string]string              // originalURL -> shortID
	IdxUserID      map[string]map[string]struct{} // userID -> shortID -> struct{}
}

func NewFileRepo(path string) (*fileRepo, error) {
	r := &fileRepo{
		path:           path,
		URLRecords:     make(map[string]*model.URLRecord),
		IdxOriginalURL: make(map[string]string),
		IdxUserID:      make(map[string]map[string]struct{}),
	}

	if err := r.loadFromFile(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *fileRepo) Save(userID, originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShort, exists := r.IdxOriginalURL[originalURL]; exists {
		if err := r.saveToFile(); err != nil {
			return "", err
		}

		return existingShort, errx.ErrConflict
	}

	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	r.URLRecords[shortID] = &model.URLRecord{OriginalURL: originalURL, UserID: userID}
	r.IdxOriginalURL[originalURL] = shortID

	if _, ok := r.IdxUserID[userID]; !ok {
		r.IdxUserID[userID] = make(map[string]struct{})
	}
	r.IdxUserID[userID][shortID] = struct{}{}

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

	record, ok := r.URLRecords[shortID]
	if !ok {
		return "", errx.ErrNotFound
	}

	if record.IsDeleted {
		return "", errx.ErrGone
	}

	return record.OriginalURL, nil
}

func (r *fileRepo) FindAll(userID string) ([]model.UserURLsResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var res []model.UserURLsResponse
	for shortID := range r.IdxUserID[userID] {
		res = append(res, model.UserURLsResponse{ShortURL: shortID, OriginalURL: r.URLRecords[shortID].OriginalURL})
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
	return encoder.Encode(&r.URLRecords)
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
	if err := decoder.Decode(&r.URLRecords); err != nil {
		return err
	}

	for k, v := range r.URLRecords {
		r.IdxOriginalURL[v.OriginalURL] = k

		if _, ok := r.IdxUserID[v.UserID]; !ok {
			r.IdxUserID[v.UserID] = make(map[string]struct{})
		}

		r.IdxUserID[v.UserID][k] = struct{}{}
	}

	return nil
}

func (r *fileRepo) DeleteURLs(userID string, shortIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.IdxUserID[userID]; !ok {
		return nil
	}

	for _, shortID := range shortIDs {
		if _, ok := r.IdxUserID[userID][shortID]; ok {
			r.URLRecords[shortID].IsDeleted = true
		}
	}

	return r.saveToFile()
}
