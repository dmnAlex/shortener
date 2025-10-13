package main

import (
	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
)

func main() {
	cfg := config.New()
	repo := repository.NewInMemoryRepo()
	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg)
	router := newRouter(handler)

	router.Run(cfg.LaunchAddress.String())
}
