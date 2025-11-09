package main

import (
	"log"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if err = logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("init logger error: %v", err)
	}

	var pgRepo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		pgRepo, err = repository.NewPostgresRepo(cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("pg repo error: %v", err)
		}
	}

	repo, err := repository.NewFileRepo(cfg.FileStoragePath)
	if err != nil {
		log.Fatalf("repo error: %v", err)
	}
	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg)
	router := newRouter(handler, pgRepo)

	if err := router.Run(cfg.LaunchAddress.String()); err != nil {
		log.Fatalf("router run error: %v", err)
	}
}
