package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/gin-gonic/gin"
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
			name: "successful shortening",
			body: "https://example.com",
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus:      http.StatusCreated,
			expectedContentType: "text/plain; charset=utf-8",
			expectedBody:        fmt.Sprintf("http://%s:%s/EwHXdJfB", host, port),
		},
		{
			name:           "empty body",
			body:           "",
			shortenFunc:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name: "service error",
			body: "https://example.com",
			shortenFunc: func(url string) (string, error) {
				return "", errors.New("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errx.ErrInternalError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockService{
				shortenFunc: tt.shortenFunc,
			}

			h := NewShortenerHandler(mockService, &config.Config{Host: host, Port: port})
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			h.HandleShorten(c)
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
		shortID             string
		expandFunc          func(string) (string, error)
		expectedStatus      int
		expectedContentType string
		expectedLocation    string
		expectedBody        string
	}{
		{
			name:    "successful redirect",
			shortID: "EwHXdJfB",
			expandFunc: func(shortID string) (string, error) {
				return "https://example.com", nil
			},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:    "not found",
			shortID: "nonexistent",
			expandFunc: func(shortID string) (string, error) {
				return "", errx.ErrNotFound
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   errx.ErrNotFound.Error(),
		},
		{
			name:           "empty id",
			shortID:        "",
			expandFunc:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name:    "service error",
			shortID: "EwHXdJfB",
			expandFunc: func(shortID string) (string, error) {
				return "", errors.New("unexpected error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   errx.ErrInternalError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockService{
				expandFunc: tt.expandFunc,
			}

			h := NewShortenerHandler(mockService, &config.Config{Host: host, Port: port})
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.shortID, nil)
			c.Params = []gin.Param{{Key: "id", Value: tt.shortID}}

			h.HandleRedirect(c)
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
