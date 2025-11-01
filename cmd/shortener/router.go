package main

import (
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/gin-gonic/gin"
)

func newRouter(h *handler.ShortenerHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	r.POST("", logger.RequestLogger(h.HandleShorten))
	r.GET("/:id", logger.RequestLogger(h.HandleRedirect))
	r.POST("/api/shorten", logger.RequestLogger(h.HandleAPIShorten))

	return r
}
