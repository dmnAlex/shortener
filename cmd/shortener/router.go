package main

import (
	"github.com/dmnAlex/shortener/internal/gzip"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/gin-gonic/gin"
)

func newRouter(h *handler.ShortenerHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.LoggerMiddleware())
	r.Use(gzip.GzipDecompressMiddleware())
	r.Use(gzip.GzipCompressMiddleware())

	r.POST("", h.HandleShorten)
	r.GET("/:id", h.HandleRedirect)
	r.POST("/api/shorten", h.HandleAPIShorten)
	r.GET("/ping", h.HandlePing)

	return r
}
