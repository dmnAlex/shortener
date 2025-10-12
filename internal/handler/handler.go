package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dmnAlex/shortener/internal/config"
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

func (h *ShortenerHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("", h.HandleShorten)
	router.GET("/:id", h.HandleRedirect)
}

func (h *ShortenerHandler) HandleShorten(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		c.String(http.StatusBadRequest, errx.ErrBadRequest.Error())
		return
	}

	originalURL := string(body)
	shortID, err := h.service.Shorten(originalURL)
	if err != nil {
		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	shortURL := fmt.Sprintf("http://%s/%s", h.config.GetAddress(), shortID)
	c.String(http.StatusCreated, shortURL)
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

		c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
		return
	}

	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}
