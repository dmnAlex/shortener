package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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
	port = 8080
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

func TestShortenerHandler_Shorten(t *testing.T) {
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
			expectedBody:        fmt.Sprintf("http://%s:%d/EwHXdJfB", host, port),
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

			h := NewShortenerHandler(mockService, &config.Config{
				LaunchAddress:  config.Address{Host: host, Port: port},
				ShortenAddress: fmt.Sprintf("http://%s:%d", host, port),
			})
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

func TestShortenerHandler_ApiShorten(t *testing.T) {
	tests := []struct {
		name                string
		body                string
		shortenFunc         func(string) (string, error)
		expectedStatus      int
		expectedContentType string
		expectedBody        string
	}{
		{
			name: "successful shortening with JSON",
			body: `{"url": "https://example.com"}`,
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus:      http.StatusCreated,
			expectedContentType: "application/json",
			expectedBody:        `{"result":"http://` + host + `:` + strconv.Itoa(port) + `/EwHXdJfB"}`,
		},
		{
			name:           "empty body",
			body:           "",
			shortenFunc:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name: "invalid JSON",
			body: `{"url": "https://example.com"`,
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name: "missing url field",
			body: `{"not_url": "https://example.com"}`,
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name: "empty url field",
			body: `{"url": ""}`,
			shortenFunc: func(url string) (string, error) {
				return "EwHXdJfB", nil
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   errx.ErrBadRequest.Error(),
		},
		{
			name: "service error",
			body: `{"url": "https://example.com"}`,
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

			h := NewShortenerHandler(mockService, &config.Config{
				LaunchAddress:  config.Address{Host: host, Port: port},
				ShortenAddress: fmt.Sprintf("http://%s:%d", host, port),
			})
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.HandleApiShorten(c)
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

				if tt.expectedContentType == "application/json" {
					var expectedJSON, actualJSON interface{}

					err = json.Unmarshal([]byte(tt.expectedBody), &expectedJSON)
					require.NoError(t, err)

					err = json.Unmarshal(body, &actualJSON)
					require.NoError(t, err)

					assert.Equal(t, expectedJSON, actualJSON)
				} else {
					assert.Equal(t, tt.expectedBody, string(body))
				}
			}
		})
	}
}

func TestShortenerHandler_Redirect(t *testing.T) {
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

			h := NewShortenerHandler(mockService, &config.Config{
				LaunchAddress:  config.Address{Host: host, Port: port},
				ShortenAddress: fmt.Sprintf("http://%s:%d", host, port),
			})
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
