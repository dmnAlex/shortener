package service

import (
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
)

type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(shortID string) (string, error)
	ShortenBatch(batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)
	Ping() error
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(originalURL string) (string, error) {
	return s.repo.Save(originalURL)
}

func (s *urlService) ShortenBatch(batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	return s.repo.SaveBatch(batch)
}

func (s *urlService) Expand(shortID string) (string, error) {
	return s.repo.Find(shortID)
}

func (s *urlService) Ping() error {
	return s.repo.Ping()
}
