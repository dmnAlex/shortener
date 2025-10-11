package main

import (
	"log"
	"net/http"

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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", handler.HandleShorten)
	mux.HandleFunc("GET /{id}", handler.HandleRedirect)

	log.Printf("server is listening on: %s", cfg.GetAddress())
	log.Fatal(http.ListenAndServe(cfg.GetAddress(), mux))
}
