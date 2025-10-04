package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type CacheItem struct {
	Data      []byte
	Headers   map[string]string
	ExpiresAt time.Time
}

type InMemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{
		items: make(map[string]*CacheItem),
	}

	go cache.cleanup()

	return cache
}

func (c *InMemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *InMemoryCache) Get(key string) (*CacheItem, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item, true
}

func (c *InMemoryCache) Set(key string, data []byte, headers map[string]string, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Data:      data,
		Headers:   headers,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *InMemoryCache) Delete(key string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
	return nil
}

func (c *InMemoryCache) Clear() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
	return nil
}

func (c *InMemoryCache) InvalidatePattern(pattern string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key := range c.items {
		if strings.Contains(key, pattern) {
			delete(c.items, key)
		}
	}
	return nil
}

func (c *InMemoryCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return CacheStats{
		KeyCount:  int64(len(c.items)),
		CacheType: "in_memory",
		HitRatio:  0.0, // In-memory cache doesn't track hits/misses currently
	}
}

func (c *InMemoryCache) ResetStats() error {
	// In-memory cache doesn't track detailed stats, so just return nil
	return nil
}

var globalCache = NewInMemoryCache()

func generateCacheKey(c *gin.Context) string {
	url := c.Request.URL.String()
	method := c.Request.Method

	hash := md5.Sum([]byte(method + ":" + url))
	return hex.EncodeToString(hash[:])
}

func CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		cacheKey := generateCacheKey(c)

		if cachedItem, exists := globalCache.Get(cacheKey); exists {
			for key, value := range cachedItem.Headers {
				c.Header(key, value)
			}
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, cachedItem.Headers["Content-Type"], cachedItem.Data)
			c.Abort()
			return
		}

		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
			headers:        make(map[string]string),
		}
		c.Writer = writer

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			for key, values := range c.Writer.Header() {
				if len(values) > 0 {
					writer.headers[key] = values[0]
				}
			}

			if err := globalCache.Set(cacheKey, writer.body, writer.headers, ttl); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
			c.Header("X-Cache", "MISS")
		}
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body    []byte
	headers map[string]string
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func InvalidateCachePattern(pattern string) {
	globalCache.InvalidatePattern(pattern)
}

func ClearCache() {
	globalCache.Clear()
}
