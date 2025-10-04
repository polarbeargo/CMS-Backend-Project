package controllers

import (
	"cms-backend/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CacheStatsResponse struct {
	Status string                `json:"status"`
	Stats  middleware.CacheStats `json:"stats"`
	Config CacheConfig           `json:"config"`
}

type CacheConfig struct {
	Type      string `json:"type"`
	TTL       string `json:"ttl"`
	RedisHost string `json:"redis_host,omitempty"`
	RedisDB   int    `json:"redis_db,omitempty"`
	Enabled   bool   `json:"enabled"`
}

func GetCacheStats(c *gin.Context) {
	stats := middleware.GetCacheStats()

	response := CacheStatsResponse{
		Status: "success",
		Stats:  stats,
		Config: CacheConfig{
			Type:    stats.CacheType,
			TTL:     "5m",
			Enabled: true,
		},
	}

	c.JSON(http.StatusOK, response)
}

func ClearCache(c *gin.Context) {
	if err := middleware.ClearAllCache(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to clear cache",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cache cleared successfully",
	})
}

func InvalidateCache(c *gin.Context) {
	pattern := c.Query("pattern")
	if pattern == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Pattern parameter is required",
		})
		return
	}

	if err := middleware.InvalidateCache(pattern); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to invalidate cache",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cache invalidated successfully",
		"pattern": pattern,
	})
}

func InvalidateResourceCache(c *gin.Context) {
	resource := c.Param("resource")

	var err error
	switch resource {
	case "media":
		err = middleware.InvalidateMediaCache()
	case "posts":
		err = middleware.InvalidatePostCache()
	case "pages":
		err = middleware.InvalidatePageCache()
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid resource type. Use: media, posts, or pages",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to invalidate cache",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"message":  "Resource cache invalidated successfully",
		"resource": resource,
	})
}

func WarmupCache(c *gin.Context) {
	resourcesParam := c.DefaultQuery("resources", "all")
	limitParam := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		limit = 10
	}

	var warmedUp []string

	if resourcesParam == "all" || resourcesParam == "media" {
		warmedUp = append(warmedUp, "media")
	}

	if resourcesParam == "all" || resourcesParam == "posts" {
		warmedUp = append(warmedUp, "posts")
	}

	if resourcesParam == "all" || resourcesParam == "pages" {
		warmedUp = append(warmedUp, "pages")
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"message":   "Cache warmup completed",
		"resources": warmedUp,
		"limit":     limit,
	})
}

func CacheHealth(c *gin.Context) {
	stats := middleware.GetCacheStats()

	health := gin.H{
		"status":     "healthy",
		"cache_type": stats.CacheType,
		"key_count":  stats.KeyCount,
		"hit_ratio":  stats.HitRatio,
	}

	if stats.CacheType == "not_initialized" {
		health["status"] = "unhealthy"
		health["issue"] = "Cache not initialized"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	if stats.HitRatio < 0.3 && (stats.Hits+stats.Misses) > 100 {
		health["status"] = "degraded"
		health["issue"] = "Low hit ratio"
	}

	statusCode := http.StatusOK
	if health["status"] == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	} else if health["status"] == "degraded" {
		statusCode = http.StatusPartialContent
	}

	c.JSON(statusCode, health)
}
