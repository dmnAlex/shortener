package main

import (
	"fmt"
	"net/http"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/gzip"
	"github.com/dmnAlex/shortener/internal/handler"
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	authTokenName = "auth_token"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string `json:"user_id"`
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(authTokenName)
		var claims *Claims
		if err != nil || cookie == "" {
			claims = &Claims{
				UserID: uuid.NewString(),
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedToken, err := token.SignedString([]byte(cfg.JWTSecret))
			if err != nil {
				c.String(http.StatusInternalServerError, "internal error")
				c.Abort()
				return
			}
			c.SetCookie(authTokenName, signedToken, 0, "/", "", false, true)
		} else {
			claims = &Claims{}
			tkn, err := jwt.ParseWithClaims(cookie, claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !tkn.Valid || claims.UserID == "" {
				c.String(http.StatusUnauthorized, "unauthorized")
				c.Abort()
				return
			}
		}
		c.Set("caller", &model.Caller{UserID: claims.UserID})
		c.Next()
	}
}

func newRouter(h *handler.ShortenerHandler, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.LoggerMiddleware())
	r.Use(gzip.GzipDecompressMiddleware())
	r.Use(gzip.GzipCompressMiddleware())
	r.Use(AuthMiddleware(cfg))
	r.POST("", h.HandleShorten)
	r.GET("/:id", h.HandleRedirect)
	r.POST("/api/shorten", h.HandleAPIShorten)
	r.POST("/api/shorten/batch", h.HandleAPIShortenBatch)
	r.GET("/api/user/urls", h.HandleAPIUserURLs)
	r.GET("/ping", h.HandlePing)
	r.DELETE("/api/user/urls", h.HandleAPIDeleteURLs)

	return r
}
