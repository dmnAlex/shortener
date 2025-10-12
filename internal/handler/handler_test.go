package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost"
	port = "8080"
)

type mockService struct {
	shortenFunc func(url string) (string, error)
	expandFunc  func(shortID string) (string, error)
}

func (m *mockService) Shorten(url string) (string, error) {
	if m.shortenFunc == nil {
		return "", errx.ErrInternalError
	}

	return m.shortenFunc(url)
}

func (m *mockService) Expand(shortID string) (string, error) {
	if m.expandFunc == nil {
		return "", errx.ErrInternalError
	}

	return m.expandFunc(shortID)
}

func TestShortenerHanler_Shorten(t *testing.T) {
	tests := []struct {
		name                string
		method              string
		body                string
		shortenFunc         func(string) (string, error)
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name:   "successful shortening",
			method: http.MethodPost,
			body:   "https://example.com",
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus:      http.StatusCreated,
			expectedContentType: "text/plain",
			expectedBody:        fmt.Sprintf("http://%s:%s/EwHXdJfB", host, port),
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			body:           "https://example.com",
			shortenFunc:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   errx.ErrMethodNotAllowed.Error() + "\n",
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			body:           "",
			shortenFunc:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error() + "\n",
		},
		{
			name:   "internal error",
			method: http.MethodPost,
			body:   "https://example.com",
			shortenFunc: func(url string) (string, error) {
				return "", errors.New("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errx.ErrInternalError.Error() + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockService{
				shortenFunc: tt.shortenFunc,
			}

			h := NewShortenerHandler(mockService, &config.Config{Host: host, Port: port})
			r := httptest.NewRequest(tt.method, "/", bytes.NewBufferString(tt.body))
			w := httptest.NewRecorder()
			h.HandleShorten(w, r)
			res := w.Result()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, res.Header.Get("Content-Type"))
			}

			if tt.expectedBody != "" {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}

func TestShortenerHanler_Redirect(t *testing.T) {
	tests := []struct {
		name                string
		method              string
		path                string
		expandFunc          func(string) (string, error)
		expectedStatus      int
		expectedContentType string
		expectedLocation    string
		expectedBody        string
	}{
		{
			name:   "successful redirect",
			method: http.MethodGet,
			path:   "/EwHXdJfB",
			expandFunc: func(shortID string) (string, error) {
				return "https://example.com", nil
			},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:           "wrong method",
			method:         http.MethodPost,
			path:           "/EwHXdJfB",
			expandFunc:     nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   errx.ErrMethodNotAllowed.Error() + "\n",
		},
		{
			name:   "not found",
			method: http.MethodGet,
			path:   "/nonexistent",
			expandFunc: func(shortID string) (string, error) {
				return "", errx.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   errx.ErrNotFound.Error() + "\n",
		},
		{
			name:   "internal error",
			method: http.MethodGet,
			path:   "/EwHXdJfB",
			expandFunc: func(shortID string) (string, error) {
				return "", errors.New("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errx.ErrInternalError.Error() + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockService{
				expandFunc: tt.expandFunc,
			}

			h := NewShortenerHandler(mockService, &config.Config{Host: host, Port: port})
			r := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			h.HandleRedirect(w, r)
			res := w.Result()

			require.Equal(t, tt.expectedStatus, res.StatusCode)

			if tt.expectedContentType != "" {
				assert.Equal(t, tt.expectedContentType, res.Header.Get("Content-Type"))
			}

			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, res.Header.Get("Location"))
			}

			if tt.expectedBody != "" {
				body, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				err = res.Body.Close()
				require.NoError(t, err)

				assert.Equal(t, tt.expectedBody, string(body))
			}
		})
	}
}
