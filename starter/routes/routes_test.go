package routes

import (
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB() (*gorm.DB, sqlmock.Sqlmock, error) {

	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	if err != nil {
		return nil, nil, err
	}

	return gormDB, mock, nil
}

func setupTestDB() *gorm.DB {
	db, _, _ := setupMockDB()
	return db
}

func TestInitializeRoutes(t *testing.T) {

	gin.SetMode(gin.TestMode)

	t.Run("RouteInitialization", func(t *testing.T) {

		router := gin.New()
		db := setupTestDB()

		InitializeRoutes(router, db)

		routes := router.Routes()

		assert.True(t, len(routes) > 0, "Routes should be registered")

		routePaths := make([]string, len(routes))
		for i, route := range routes {
			routePaths[i] = route.Path
		}

		expectedRoutes := []string{
			"/api/v1/pages",
			"/api/v1/pages/:id",
			"/api/v1/posts",
			"/api/v1/posts/:id",
			"/api/v1/media",
			"/api/v1/media/:id",
		}

		for _, expectedRoute := range expectedRoutes {
			assert.Contains(t, routePaths, expectedRoute,
				"Route %s should be registered", expectedRoute)
		}
	})

	t.Run("RouteMethodMapping", func(t *testing.T) {
		router := gin.New()
		db := setupTestDB()
		InitializeRoutes(router, db)

		routes := router.Routes()
		methodCount := make(map[string]int)

		for _, route := range routes {
			methodCount[route.Method]++
		}

		assert.True(t, methodCount["GET"] > 0, "Should have GET routes")
		assert.True(t, methodCount["POST"] > 0, "Should have POST routes")
		assert.True(t, methodCount["PUT"] > 0, "Should have PUT routes")
		assert.True(t, methodCount["DELETE"] > 0, "Should have DELETE routes")
	})

	t.Run("MiddlewareConfiguration", func(t *testing.T) {
		router := gin.New()
		db := setupTestDB()

		router.Use(func(c *gin.Context) {
			c.Set("db", db)
			c.Next()
		})

		router.GET("/test", func(c *gin.Context) {
			dbFromContext, exists := c.Get("db")
			if exists && dbFromContext != nil {
				c.JSON(200, gin.H{"status": "ok"})
			} else {
				c.JSON(500, gin.H{"status": "error"})
			}
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"ok"`)
	})

	t.Run("APIVersioning", func(t *testing.T) {
		router := gin.New()
		db := setupTestDB()

		InitializeRoutes(router, db)

		routes := router.Routes()
		apiV1Routes := 0

		for _, route := range routes {
			if len(route.Path) >= 7 && route.Path[:7] == "/api/v1" {
				apiV1Routes++
			}
		}

		assert.True(t, apiV1Routes > 0, "Should have API v1 routes")
	})

	t.Run("RouteGroupStructure", func(t *testing.T) {
		router := gin.New()
		db := setupTestDB()

		InitializeRoutes(router, db)

		routes := router.Routes()
		resourceRoutes := map[string]int{
			"pages": 0,
			"posts": 0,
			"media": 0,
		}

		for _, route := range routes {
			if contains(route.Path, "pages") {
				resourceRoutes["pages"]++
			} else if contains(route.Path, "posts") {
				resourceRoutes["posts"]++
			} else if contains(route.Path, "media") {
				resourceRoutes["media"]++
			}
		}

		for resource, count := range resourceRoutes {
			assert.True(t, count >= 4,
				"Resource %s should have at least 4 routes (CRUD)", resource)
		}
	})
}

func TestDatabaseMiddleware(t *testing.T) {
	t.Run("DatabaseContextInjection", func(t *testing.T) {
		router := gin.New()
		db := setupTestDB()

		router.Use(func(c *gin.Context) {
			c.Set("db", db)
			c.Next()
		})

		router.GET("/test", func(c *gin.Context) {
			dbFromContext, exists := c.Get("db")
			if exists && dbFromContext != nil {
				c.JSON(200, gin.H{"db_injected": true})
			} else {
				c.JSON(500, gin.H{"db_injected": false})
			}
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), `"db_injected":true`)
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		s[max(0, len(s)-len(substr)):] == substr ||
		s[:min(len(substr), len(s))] == substr ||
		(len(s) > len(substr) &&
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
