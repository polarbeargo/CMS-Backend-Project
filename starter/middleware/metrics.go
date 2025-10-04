package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		if duration > time.Second {
			log.Printf("SLOW REQUEST: %s %s - %v - Status: %d",
				method, path, duration, statusCode)
		}
		c.Header("X-Response-Time", duration.String())
	}
}

type DatabaseMetrics struct {
	ActiveConnections int
	IdleConnections   int
	TotalConnections  int
}

func GetDatabaseMetrics() *DatabaseMetrics {
	return &DatabaseMetrics{
		ActiveConnections: 0,
		IdleConnections:   0,
		TotalConnections:  0,
	}
}
