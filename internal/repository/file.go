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
	mu      sync.RWMutex
	file    *os.File
	encoder *json.Encoder

	URLRecords     map[string]*model.URLRecord    // shortID -> shortID -> *model.URLRecord
	IdxOriginalURL map[string]string              // originalURL -> shortID
	IdxUserID      map[string]map[string]struct{} // userID -> shortID -> struct{}
}

func NewFileRepo(path string) (*fileRepo, error) {
	r := &fileRepo{
		URLRecords:     make(map[string]*model.URLRecord),
		IdxOriginalURL: make(map[string]string),
		IdxUserID:      make(map[string]map[string]struct{}),
	}

	if path != "" {
		if err := r.loadFromFile(path); err != nil {
			return nil, err
		}

		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		r.file = file
		r.encoder = json.NewEncoder(file)
	}

	return r, nil
}

func (r *fileRepo) Save(userID, originalURL string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingShort, exists := r.IdxOriginalURL[originalURL]; exists {
		return existingShort, errx.ErrConflict
	}

	shortID, err := utils.GenerateShortID()
	if err != nil {
		return "", err
	}

	record := &model.URLRecord{OriginalURL: originalURL, UserID: userID}
	entry := model.FileEntry{ShortID: shortID, URLRecord: *record}

	if r.encoder != nil {
		if err := r.encoder.Encode(&entry); err != nil {
			return "", err
		}
	}

	r.updateMemory(shortID, record)

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
		if !r.URLRecords[shortID].IsDeleted {
			res = append(res, model.UserURLsResponse{ShortURL: shortID, OriginalURL: r.URLRecords[shortID].OriginalURL})
		}
	}

	return res, nil
}

func (r *fileRepo) Ping() error {
	return nil
}

func (r *fileRepo) Close() error {
	if r.file != nil {
		return r.file.Close()
	}

	return nil
}

func (r *fileRepo) loadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var entry model.FileEntry
		if err := decoder.Decode(&entry); err != nil {
			return err
		}

		r.updateMemory(entry.ShortID, &entry.URLRecord)
	}

	return nil
}

func (r *fileRepo) updateMemory(shortID string, record *model.URLRecord) {
	r.URLRecords[shortID] = record

	r.IdxOriginalURL[record.OriginalURL] = shortID
	if _, ok := r.IdxUserID[record.UserID]; !ok {
		r.IdxUserID[record.UserID] = make(map[string]struct{})
	}
	r.IdxUserID[record.UserID][shortID] = struct{}{}
}

func (r *fileRepo) DeleteURLs(userID string, shortIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.IdxUserID[userID]; !ok {
		return nil
	}

	for _, shortID := range shortIDs {
		if _, ok := r.IdxUserID[userID][shortID]; ok {
			record := r.URLRecords[shortID]
			record.IsDeleted = true

			if r.encoder != nil {
				entry := model.FileEntry{ShortID: shortID, URLRecord: *record}
				if err := r.encoder.Encode(&entry); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
