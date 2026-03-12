package middleware

import (
	"net"
	"net/http"

	"github.com/dmnAlex/shortener/internal/config"
	"github.com/gin-gonic/gin"
)

// SubnetCheck проверяет что значение заголовка X-Real-IP соответствует доверенной подсети.
func SubnetCheck(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.TrustedSubnet == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		_, network, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		clientIP := net.ParseIP(c.GetHeader("X-Real-IP"))
		if clientIP == nil || !network.Contains(clientIP) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
