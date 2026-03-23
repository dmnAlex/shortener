// Package service предоставляет бизнес-логику для работы с URL.
package service

import (
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
)

// URLService определяет контракт сервиса для работы с URL.
type URLService interface {
	// Shorten создает короткую ссылку для оригинального URL пользователя.
	// Возвращает короткий ID или ошибку ErrConflict если URL уже существует.
	Shorten(userID, originalURL string) (string, error)

	// ShortenBatch создает несколько коротких ссылок пакетно.
	// Сохраняет соответствие correlation_id из запроса.
	ShortenBatch(userID string, batch []model.ShortenBatchRequest) ([]model.ShortenBatchResponse, error)

	// Expand возвращает оригинальный URL по короткому ID.
	// Возвращает ErrNotFound если ссылка не найдена, ErrGone если удалена.
	Expand(shortID string) (string, error)

	// UserURLs возвращает все активные ссылки пользователя.
	// Возвращает пустой срез если ссылок нет.
	UserURLs(userID string) ([]model.UserURLsResponse, error)

	// Ping проверяет доступность хранилища.
	Ping() error

	// DeleteURLs помечает ссылки пользователя как удаленные.
	// Удаление выполняется асинхронно.
	DeleteURLs(userID string, shortIDs []string) error

	// Stats возвращает количество URL и пользователей
	Stats() (int, int, error)
}

type urlService struct {
	repo repository.URLRepository
}

// NewURLService создает новый экземпляр URLService с указанным репозиторием.
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

func (s *urlService) Stats() (int, int, error) {
	return s.repo.Stats()
}
