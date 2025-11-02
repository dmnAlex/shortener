package logger

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Init(level string) error {
	var err error
	cfg := zap.NewProductionConfig()
	if cfg.Level, err = zap.ParseAtomicLevel(level); err != nil {
		return err
	}

	if Log, err = cfg.Build(); err != nil {
		return err
	}

	return nil
}

type responseData struct {
	status int
	size   int
}

type customWriter struct {
	gin.ResponseWriter
	responseData *responseData
}

func (cw *customWriter) Write(b []byte) (int, error) {
	size, err := cw.ResponseWriter.Write(b)
	cw.responseData.size += size
	return size, err
}

func (cw *customWriter) WriteHeader(statusCode int) {
	cw.ResponseWriter.WriteHeader(statusCode)
	cw.responseData.status = statusCode
}

func LoggerMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now
		rd := &responseData{status: 0, size: 0}
		cw := &customWriter{ResponseWriter: c.Writer, responseData: rd}
		c.Writer = cw

		c.Next()

		Log.Info("got incoming HTTP request",
			zap.String("uri", c.Request.URL.String()),
			zap.String("method", c.Request.Method),
			zap.String("duration", time.Since(start()).String()),
			zap.String("status", strconv.Itoa(rd.status)),
			zap.String("size", strconv.Itoa(rd.size)),
		)
	})
}
