package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/dmnAlex/shortener/internal/repository"
	"github.com/dmnAlex/shortener/internal/router"
	"github.com/dmnAlex/shortener/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func ExampleShortenerHandler_HandleShorten() {
	// Создаем тестовый репозиторий
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	// Создаем сервис и хендлер
	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	// Создаем тестовый контекст Gin
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))

	// Вызываем хендлер
	h.HandleShorten(c)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 201
}

func ExampleShortenerHandler_HandleAPIShorten() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})

	// JSON запрос
	jsonBody := `{"url": "https://example.org"}`
	c.Request = httptest.NewRequest("POST", "/api/shorten", strings.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleAPIShorten(c)

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Result contains: %v\n", strings.Contains(resp["result"], "http://localhost:8080/"))
	// Output:
	// Status: 201
	// Result contains: true
}

func ExampleShortenerHandler_HandleRedirect() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	// Сначала создаем короткую ссылку
	shortID, _ := repo.Save("test-user", "https://example.com")

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})
	c.Request = httptest.NewRequest("GET", "/"+shortID, nil)
	c.Params = []gin.Param{{Key: "id", Value: shortID}}

	h.HandleRedirect(c)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))
	// Output:
	// Status: 307
	// Location: https://example.com
}

func ExampleShortenerHandler_HandleAPIShortenBatch() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})

	// Пакетный запрос
	batchBody := `[
        {"correlation_id": "1", "original_url": "https://example1.com"},
        {"correlation_id": "2", "original_url": "https://example2.com"}
    ]`
	c.Request = httptest.NewRequest("POST", "/api/shorten/batch", strings.NewReader(batchBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleAPIShortenBatch(c)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 201
}

func ExampleShortenerHandler_HandleAPIUserURLs() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	// Создаем несколько ссылок
	repo.Save("test-user", "https://example1.com")
	repo.Save("test-user", "https://example2.com")

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})
	c.Request = httptest.NewRequest("GET", "/api/user/urls", nil)

	h.HandleAPIUserURLs(c)

	var urls []map[string]string
	json.Unmarshal(w.Body.Bytes(), &urls)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("URLs count: %d\n", len(urls))
	// Output:
	// Status: 200
	// URLs count: 2
}

func ExampleShortenerHandler_HandleAPIDeleteURLs() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	// Создаем ссылку
	shortID, _ := repo.Save("test-user", "https://example.com")

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
	}
	h := handler.NewShortenerHandler(svc, cfg, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("caller", &model.Caller{UserID: "test-user"})

	// Запрос на удаление
	deleteBody := fmt.Sprintf(`["%s"]`, shortID)
	c.Request = httptest.NewRequest("DELETE", "/api/user/urls", strings.NewReader(deleteBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h.HandleAPIDeleteURLs(c)

	fmt.Printf("Status: %d\n", w.Code)
	// Output:
	// Status: 200
}

func ExampleShortenerHandler_HandleInternalStats() {
	repo, _ := repository.NewFileRepo("")
	defer repo.Close()

	repo.Save("user1", "https://example1.com")
	repo.Save("user1", "https://example2.com")
	repo.Save("user2", "https://example3.com")

	svc := service.NewURLService(repo)
	cfg := &config.Config{
		ShortenAddress: "http://localhost:8080",
		TrustedSubnet:  "127.0.0.1/32",
		JWTSecret:      "test-secret",
	}

	h := handler.NewShortenerHandler(svc, cfg, nil)

	router := router.New(h, cfg)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/internal/stats", nil)
	req.Header.Set("X-Real-IP", "127.0.0.1")

	token := createTestToken(cfg.JWTSecret, "test-user")
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: token,
	})

	router.ServeHTTP(w, req)

	var stats model.StatsResponse
	json.Unmarshal(w.Body.Bytes(), &stats)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("URLs: %d, Users: %d\n", stats.URLs, stats.Users)
	// Output:
	// Status: 200
	// URLs: 3, Users: 2
}

func createTestToken(secret, userID string) string {
	claims := &model.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}
