package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type ShortenerHandler struct {
	service service.URLService
	config  *config.Config
}

func NewShortenerHandler(s service.URLService, cfg *config.Config) *ShortenerHandler {
	return &ShortenerHandler{
		service: s,
		config:  cfg,
	}
}

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

	c.String(status, shortURL)
}

func (h *ShortenerHandler) HandleAPIShorten(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	var req model.ShortenRequest

	if err := json.Unmarshal(body, &req); err != nil || req.URL == "" {
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

	res, err := json.Marshal(model.ShortenResponse{Result: shortURL})
	if err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.Data(status, "application/json", res)
}

func (h *ShortenerHandler) shortenURL(c *gin.Context, url string) (string, error) {
	shortID, err := h.service.Shorten(c, url)
	return fmt.Sprintf("%s/%s", h.config.ShortenAddress, shortID), err
}

func (h *ShortenerHandler) HandleAPIShortenBatch(c *gin.Context) {
	var req []model.ShortenBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	res, err := h.service.ShortenBatch(c, req)
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

func (h *ShortenerHandler) HandleAPIUserURLs(c *gin.Context) {
	res, err := h.service.UserURLs(c)
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

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

func (h *ShortenerHandler) HandlePing(c *gin.Context) {
	if err := h.service.Ping(); err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.String(http.StatusOK, "OK")
}

func (h *ShortenerHandler) HandleAPIDeleteURLs(c *gin.Context) {
	var shortIDs []string
	if err := c.ShouldBindJSON(&shortIDs); err != nil {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	if err := h.service.DeleteURLs(c, shortIDs); err != nil {
		logger.Log.Error(err.Error())
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.Status(http.StatusAccepted)
}
