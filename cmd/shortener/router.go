package main

import (
	"net/http"

	"github.com/dmnAlex/shortener/internal/gzip"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/gin-gonic/gin"
)

func newRouter(h *handler.ShortenerHandler, pgRepo repository.URLRepository) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.LoggerMiddleware())
	r.Use(gzip.GzipDecompressMiddleware())
	r.Use(gzip.GzipCompressMiddleware())

	r.POST("", h.HandleShorten)
	r.GET("/:id", h.HandleRedirect)
	r.POST("/api/shorten", h.HandleAPIShorten)
	r.GET("/ping", func(c *gin.Context) {
		if pgRepo == nil {
			c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
			return
		}

		if err := pgRepo.Ping(); err != nil {
			c.String(http.StatusInternalServerError, errx.ErrInternalError.Error())
			return
		}

		c.String(http.StatusOK, "OK")
	})

	return r
}
