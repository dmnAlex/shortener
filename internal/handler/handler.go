package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/dmnAlex/shortener/internal/service"
)

type ShortenerHandler struct {
	service service.URLService
	config  *config.Config
}

func NewShortenerHandler(s service.URLService, cfg *config.Config) *ShortenerHandler {
	return &ShortenerHandler{
		service: s,
		config:  cfg,
	}
}

func (h *ShortenerHandler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, errx.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, errx.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	originalURL := string(body)
	shortID, err := h.service.Shorten(originalURL)
	if err != nil {
		http.Error(w, errx.ErrInternalError.Error(), http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("http://%s/%s", h.config.GetAddress(), shortID)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h *ShortenerHandler) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, errx.ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		http.Error(w, errx.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.Expand(path)
	if err != nil {
		if errors.Is(err, errx.ErrNotFound) {
			http.Error(w, errx.ErrNotFound.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, errx.ErrInternalError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
