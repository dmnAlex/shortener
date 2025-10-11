package service

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/dmnAlex/shortener/internal/repository"
)

type URLService interface {
	Shorten(originalURL string) (string, error)
	Expand(shortID string) (string, error)
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(originalURL string) (string, error) {
	shortID := generateShortID()
	err := s.repo.Save(shortID, originalURL)
	return shortID, err
}

func (s *urlService) Expand(shortID string) (string, error) {
	return s.repo.Find(shortID)
}

func generateShortID() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
