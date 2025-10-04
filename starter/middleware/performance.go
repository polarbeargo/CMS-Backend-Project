package middleware

import (
	"compress/gzip"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer io.Writer
}

func (g gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if shouldSkipCompression(contentType) {
			c.Next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()
	}
}

func shouldSkipCompression(contentType string) bool {
	skipTypes := []string{
		"image/",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/x-gzip",
	}

	for _, skipType := range skipTypes {
		if strings.HasPrefix(contentType, skipType) {
			return true
		}
	}
	return false
}

func ConnectionPoolMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Connection", "keep-alive")
		c.Header("Keep-Alive", "timeout=120, max=1000")

		c.Next()
	}
}
