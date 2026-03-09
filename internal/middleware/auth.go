package middleware

import (
	"fmt"
	"net/http"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

const (
	authTokenName = "auth_token"
)

func Auth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(authTokenName)
		var claims *model.Claims
		if err != nil || cookie == "" {
			claims = &model.Claims{
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
			claims = &model.Claims{}
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
