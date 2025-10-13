package main

import (
	"log"

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

	if err := router.Run(cfg.LaunchAddress.String()); err != nil {
		log.Fatalf("router run error: %v", err)
	}
}
