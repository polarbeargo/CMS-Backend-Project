// main.go
package main

import (
	"cms-backend/models"
	"cms-backend/routes"
	"cms-backend/utils"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/time/rate"
)

// @title CMS Backend API
// @version 1.0
// @description This is a backend API for a Content Management System (CMS).
// @host localhost:8080
// @BasePath /api/v1

type ipRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

func newIPRateLimiter(r rate.Limit, b int) *ipRateLimiter {
	return &ipRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (i *ipRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()
	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.limiters[ip] = limiter
	}
	return limiter
}

func RateLimitMiddleware(limiter *ipRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}
		if !limiter.getLimiter(ip).Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "Too Many Requests"})
			return
		}
		c.Next()
	}
}

func SecureHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Set("Referrer-Policy", "no-referrer")
		c.Writer.Header().Set("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func main() {
	// Initialize database connection
	dbRes, err := utils.ConnectDB()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer utils.CloseDatabase(dbRes)

	// Get the environment variable
	env := os.Getenv("ENV")
	if env == "" {
		env = "development" // default to development if ENV is not set
	}

	// Conditionally run AutoMigrate in development environment
	if env == "development" {
		log.Println("Running AutoMigrate...")
		if err := dbRes.GormDB.AutoMigrate(&models.Page{}, &models.Post{}, &models.Media{}); err != nil {
			log.Fatalf("Failed to automigrate database: %v", err)
		}
	}

	// Set Gin mode based on environment
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(SecureHeader())

	limiter := newIPRateLimiter(rate.Every(time.Minute/100), 100)
	router.Use(RateLimitMiddleware(limiter))

	// Initialize routes
	routes.InitializeRoutes(router, dbRes.GormDB)

	// Run the server
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
