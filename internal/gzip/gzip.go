package gzip

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/dmnAlex/shortener/internal/model/errx"
	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	w *gzip.Writer
}

func newGzipWriter(w gin.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		w:              gzip.NewWriter(w),
	}
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.w.Write(data)
}

func (g *gzipWriter) Close() error {
	return g.w.Close()
}

func GzipDecompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithError(http.StatusBadRequest, errx.ErrBadRequest)
				return
			}
			defer reader.Close()

			c.Request.Body = reader
		}

		c.Next()
	}
}

func GzipCompressMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			contentType := c.GetHeader("Content-Type")
			shouldCompress := strings.Contains(contentType, "application/json") ||
				strings.Contains(contentType, "text/html")

			if shouldCompress {
				gw := newGzipWriter(c.Writer)
				defer gw.Close()
				c.Writer = gw

				c.Header("Content-Encoding", "gzip")
			}
		}

		c.Next()
	}
}
