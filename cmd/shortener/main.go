package main

import (
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.New()
	repo := repository.NewInMemoryRepo()
	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg)

	router := gin.Default()
	handler.RegisterRoutes(router)

	router.Run(cfg.GetLaunchAddress())
}
