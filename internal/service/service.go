package service

import (
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

type URLService interface {
	Shorten(c *gin.Context, originalURL string) (string, error)
	ShortenBatch(c *gin.Context, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)
	Expand(shortID string) (string, error)
	UserURLs(c *gin.Context) ([]model.UserURLsResponse, error)
	Ping() error
	DeleteURLs(c *gin.Context, shortIDs []string) error
}

type urlService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) URLService {
	return &urlService{repo: repo}
}

func (s *urlService) Shorten(c *gin.Context, originalURL string) (string, error) {
	caller := c.MustGet("caller").(*model.Caller)
	return s.repo.Save(caller.UserID, originalURL)
}

func (s *urlService) ShortenBatch(c *gin.Context, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error) {
	caller := c.MustGet("caller").(*model.Caller)
	return s.repo.SaveBatch(caller.UserID, batch)
}

func (s *urlService) UserURLs(c *gin.Context) ([]model.UserURLsResponse, error) {
	caller := c.MustGet("caller").(*model.Caller)
	return s.repo.FindAll(caller.UserID)
}

func (s *urlService) Expand(shortID string) (string, error) {
	return s.repo.Find(shortID)
}

func (s *urlService) Ping() error {
	return s.repo.Ping()
}

func (s *urlService) DeleteURLs(c *gin.Context, shortIDs []string) error {
	caller := c.MustGet("caller").(*model.Caller)
	return s.repo.DeleteURLs(caller.UserID, shortIDs)
}
