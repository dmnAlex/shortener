package main

import (
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/gin-gonic/gin"
)

func newRouter(h *handler.ShortenerHandler) *gin.Engine {
	r := gin.Default()

	r.POST("", h.HandleShorten)
	r.GET("/:id", h.HandleRedirect)

	return r
}
