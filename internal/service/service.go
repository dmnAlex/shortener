package service

import (
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
)

type URLService interface {
	Shorten(userID, originalURL string) (string, error)
	ShortenBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)
	Expand(shortID string) (string, error)
	UserURLs(userID string) ([]model.UserURLsResponse, error)
	Ping() error
	DeleteURLs(userID string, shortIDs []string) error
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(userID, originalURL string) (string, error) {
	return s.repo.Save(userID, originalURL)
}

func (s *urlService) ShortenBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	return s.repo.SaveBatch(userID, batch)
}

func (s *urlService) UserURLs(userID string) ([]model.UserURLsResponse, error) {
	return s.repo.FindAll(userID)
}

func (s *urlService) Expand(shortID string) (string, error) {
	return s.repo.Find(shortID)
}

func (s *urlService) Ping() error {
	return s.repo.Ping()
}

func (s *urlService) DeleteURLs(userID string, shortIDs []string) error {
	return s.repo.DeleteURLs(userID, shortIDs)
}
