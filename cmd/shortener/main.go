package main

import (
	"context"
	"log"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/dmnAlex/shortener/internal/storage/pg"
)

func main() {
	globalCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	if err = logger.Init(cfg.LogLevel); err != nil {
		log.Fatalf("init logger error: %v", err)
	}

	var repo repository.URLRepository
	if cfg.DatabaseDSN != "" {
		logger.Log.Info("using pg database")
		db, err := pg.New(globalCtx, cfg.DatabaseDSN, cfg.MigrationsPath)
		if err != nil {
			log.Fatalf("db error: %v", err)
		}
		defer db.Close()

		repo = repository.NewPostgresRepo(db)

	} else {
		logger.Log.Info("using file database")
		repo, err = repository.NewFileRepo(cfg.FileStoragePath)
		if err != nil {
			log.Fatalf("file repo error: %v", err)
		}
	}

	service := service.NewURLService(repo)
	handler := handler.NewShortenerHandler(service, cfg)
	router := newRouter(handler)

	if err := router.Run(cfg.LaunchAddress.String()); err != nil {
		log.Fatalf("router run error: %v", err)
	}
}
