package router

import (
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/middleware"
	"github.com/gin-gonic/gin"
)

func New(h *handler.ShortenerHandler, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.GzipDecompress())
	r.Use(middleware.GzipCompress())
	r.Use(middleware.Auth(cfg))
	r.POST("", h.HandleShorten)
	r.GET("/:id", h.HandleRedirect)
	r.POST("/api/shorten", h.HandleAPIShorten)
	r.POST("/api/shorten/batch", h.HandleAPIShortenBatch)
	r.GET("/api/user/urls", h.HandleAPIUserURLs)
	r.GET("/ping", h.HandlePing)
	r.DELETE("/api/user/urls", h.HandleAPIDeleteURLs)

	internal := r.Group("/api/internal")
	internal.Use(middleware.SubnetCheck(cfg))
	internal.GET("/stats", h.HandleInternalStats)

	return r
}
