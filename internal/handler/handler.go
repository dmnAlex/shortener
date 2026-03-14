// Package handler предоставляет HTTP-обработчики для сервиса сокращения URL.
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dmnAlex/shortener/internal/audit"
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// ShortenerHandler обрабатывает HTTP-запросы для сервиса сокращения URL.
// Предоставляет методы для создания коротких ссылок, редиректа и управления URL.
type ShortenerHandler struct {
	service  service.URLService
	config   *config.Config
	auditMgr *audit.AuditManager
}

// NewShortenerHandler создает новый экземпляр ShortenerHandler.
// Принимает сервис URL, конфигурацию и менеджер аудита (опционально).
func NewShortenerHandler(s service.URLService, cfg *config.Config, auditMgr *audit.AuditManager) *ShortenerHandler {
	return &ShortenerHandler{
		service:  s,
		config:   cfg,
		auditMgr: auditMgr,
	}
}

// HandleShorten обрабатывает POST запрос для создания короткой ссылки из plain/text тела.
// Возвращает короткий URL в формате text/plain.
// Статусы:
//   - 201 Created - URL успешно создан
//   - 409 Conflict - URL уже существует
//   - 400 Bad Request - некорректный запрос
//   - 500 Internal Server Error - внутренняя ошибка сервера
func (h *ShortenerHandler) HandleShorten(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	originalURL := string(body)
	status := http.StatusCreated
	shortURL, err := h.shortenURL(c, originalURL)
	if err != nil {
		if !errors.Is(err, errx.ErrConflict) {
			logger.Log.Error(err.Error())
			c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
			return
		}

		status = http.StatusConflict
	}

	h.notifyAudit(c, model.AuditActionShorten, originalURL)

	c.String(status, shortURL)
}

// HandleAPIShorten обрабатывает POST /api/shorten для создания короткой ссылки из JSON.
// Принимает JSON вида {"url": "https://example.com"}.
// Возвращает JSON вида {"result": "http://host:port/short_id"}.
// Статусы аналогичны HandleShorten.
func (h *ShortenerHandler) HandleAPIShorten(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	var req model.ShortenRequest

	if err = json.Unmarshal(body, &req); err != nil || req.URL == "" {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	status := http.StatusCreated
	shortURL, err := h.shortenURL(c, req.URL)
	if err != nil {
		if !errors.Is(err, errx.ErrConflict) {
			logger.Log.Error(err.Error())
			c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
			return
		}

		status = http.StatusConflict
	}

	h.notifyAudit(c, model.AuditActionShorten, req.URL)

	res, err := json.Marshal(model.ShortenResponse{Result: shortURL})
	if err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.Data(status, "application/json", res)
}

func (h *ShortenerHandler) shortenURL(c *gin.Context, url string) (string, error) {
	caller := c.MustGet(model.CallerKey).(*model.Caller)
	shortID, err := h.service.Shorten(caller.UserID, url)
	return fmt.Sprintf("%s/%s", h.config.ShortenAddress, shortID), err
}

// HandleAPIShortenBatch обрабатывает POST /api/shorten/batch для пакетного создания ссылок.
// Принимает массив объектов с correlation_id и original_url.
// Возвращает массив объектов с correlation_id и short_url.
func (h *ShortenerHandler) HandleAPIShortenBatch(c *gin.Context) {
	var req []model.ShortenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	caller := c.MustGet(model.CallerKey).(*model.Caller)
	res, err := h.service.ShortenBatch(caller.UserID, req)
	if err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	for i := range res {
		res[i].ShortURL = fmt.Sprintf("%s/%s", h.config.ShortenAddress, res[i].ShortURL)
	}

	c.JSON(http.StatusCreated, res)
}

// HandleAPIUserURLs обрабатывает GET /api/user/urls для получения всех ссылок пользователя.
// Возвращает массив объектов с short_url и original_url.
// Статус 204 No Content - если у пользователя нет ссылок.
func (h *ShortenerHandler) HandleAPIUserURLs(c *gin.Context) {
	caller := c.MustGet(model.CallerKey).(*model.Caller)
	res, err := h.service.UserURLs(caller.UserID)
	if err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	if len(res) == 0 {
		c.Status(204)
		return
	}

	for i := range res {
		res[i].ShortURL = fmt.Sprintf("%s/%s", h.config.ShortenAddress, res[i].ShortURL)
	}

	c.JSON(http.StatusOK, res)
}

// HandleRedirect обрабатывает GET /:id запрос для редиректа по короткой ссылке.
// Статусы:
//   - 307 Temporary Redirect - успешный редирект
//   - 404 Not Found - ссылка не найдена
//   - 410 Gone - ссылка удалена
//   - 400 Bad Request - некорректный ID
//   - 500 Internal Server Error - внутренняя ошибка
func (h *ShortenerHandler) HandleRedirect(c *gin.Context) {
	shortID := c.Param("id")
	if shortID == "" {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	originalURL, err := h.service.Expand(shortID)
	if err != nil {
		if errors.Is(err, errx.ErrNotFound) {
			c.String(http.StatusNotFound, errx.ErrNotFound.Error())
			return
		}

		if errors.Is(err, errx.ErrGone) {
			c.Status(http.StatusGone)
			return
		}

		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	h.notifyAudit(c, model.AuditActionFollow, originalURL)
	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

// HandlePing обрабатывает GET /ping для проверки доступности БД.
// Возвращает 200 OK если БД доступна, 500 в противном случае.
func (h *ShortenerHandler) HandlePing(c *gin.Context) {
	if err := h.service.Ping(); err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.String(http.StatusOK, "OK")
}

// HandleAPIDeleteURLs обрабатывает DELETE /api/user/urls для асинхронного удаления ссылок.
// Принимает массив коротких ID для удаления.
// Возвращает 202 Accepted - запрос принят в обработку.
func (h *ShortenerHandler) HandleAPIDeleteURLs(c *gin.Context) {
	var shortIDs []string
	if err := c.ShouldBindJSON(&shortIDs); err != nil {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	caller := c.MustGet(model.CallerKey).(*model.Caller)
	if err := h.service.DeleteURLs(caller.UserID, shortIDs); err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *ShortenerHandler) notifyAudit(c *gin.Context, action model.AuditAction, url string) {
	if h.auditMgr == nil {
		return
	}

	caller := c.MustGet(model.CallerKey).(*model.Caller)
	h.auditMgr.Notify(model.AuditEvent{
		Timestamp: time.Now().Unix(),
		Action:    action,
		UserID:    caller.UserID,
		URL:       url,
	})
}

func (h *ShortenerHandler) HandleInternalStats(c *gin.Context) {
	urls, users, err := h.service.Stats()
	if err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.JSON(http.StatusOK, model.StatsResponse{
		URLs:  urls,
		Users: users,
	})
}
