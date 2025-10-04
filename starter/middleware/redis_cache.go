package middleware

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type CacheInterface interface {
	Get(key string) (*CacheItem, bool)
	Set(key string, data []byte, headers map[string]string, ttl time.Duration) error
	Delete(key string) error
	Clear() error
	InvalidatePattern(pattern string) error
	GetStats() CacheStats
	ResetStats() error
}

type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	HitRatio    float64   `json:"hit_ratio"`
	KeyCount    int64     `json:"key_count"`
	CacheType   string    `json:"cache_type"`
	LastCleanup time.Time `json:"last_cleanup,omitempty"`
}

type RedisCache struct {
	client *redis.Client
	prefix string
	stats  CacheStats
}

func NewRedisCache() (*RedisCache, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	prefix := os.Getenv("REDIS_PREFIX")
	if prefix == "" {
		prefix = "cms_cache:"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password:     redisPassword,
		DB:           redisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s:%s (DB: %d)", redisHost, redisPort, redisDB)

	return &RedisCache{
		client: rdb,
		prefix: prefix,
		stats: CacheStats{
			CacheType: "redis",
		},
	}, nil
}

func (r *RedisCache) Get(key string) (*CacheItem, bool) {
	ctx := context.Background()
	fullKey := r.prefix + key

	data, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			r.stats.Misses++
			return nil, false
		}
		log.Printf("Redis GET error: %v", err)
		r.stats.Misses++
		return nil, false
	}

	var item CacheItem
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		log.Printf("Failed to unmarshal cache item: %v", err)
		r.stats.Misses++
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		r.Delete(key)
		r.stats.Misses++
		return nil, false
	}

	r.stats.Hits++
	return &item, true
}

func (r *RedisCache) Set(key string, data []byte, headers map[string]string, ttl time.Duration) error {
	ctx := context.Background()
	fullKey := r.prefix + key

	item := CacheItem{
		Data:      data,
		Headers:   headers,
		ExpiresAt: time.Now().Add(ttl),
	}

	jsonData, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %v", err)
	}

	if err := r.client.Set(ctx, fullKey, jsonData, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache item: %v", err)
	}

	return nil
}

func (r *RedisCache) Delete(key string) error {
	ctx := context.Background()
	fullKey := r.prefix + key

	if err := r.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("failed to delete cache item: %v", err)
	}

	return nil
}

func (r *RedisCache) Clear() error {
	ctx := context.Background()
	pattern := r.prefix + "*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys: %v", err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %v", err)
	}

	log.Printf("Cleared %d cache items", len(keys))
	return nil
}

func (r *RedisCache) InvalidatePattern(pattern string) error {
	ctx := context.Background()
	searchPattern := r.prefix + "*" + pattern + "*"

	keys, err := r.client.Keys(ctx, searchPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %v", pattern, err)
	}

	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys for pattern %s: %v", pattern, err)
	}

	log.Printf("Invalidated %d cache items matching pattern: %s", len(keys), pattern)
	return nil
}

func (r *RedisCache) GetStats() CacheStats {
	ctx := context.Background()
	pattern := r.prefix + "*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err == nil {
		r.stats.KeyCount = int64(len(keys))
	}

	total := r.stats.Hits + r.stats.Misses
	if total > 0 {
		r.stats.HitRatio = float64(r.stats.Hits) / float64(total)
	}

	return r.stats
}

func (r *RedisCache) ResetStats() error {
	r.stats.Hits = 0
	r.stats.Misses = 0
	r.stats.HitRatio = 0.0
	r.stats.KeyCount = 0
	return nil
}

type CacheManager struct {
	primary     CacheInterface
	fallback    CacheInterface
	useFallback bool
}

func NewCacheManager() *CacheManager {
	var primary CacheInterface
	var fallback CacheInterface = NewInMemoryCache()
	useFallback := false
	redisCache, err := NewRedisCache()
	if err != nil {
		log.Printf("Redis cache initialization failed, using in-memory cache: %v", err)
		primary = fallback
		useFallback = true
	} else {
		primary = redisCache
		log.Println("Using Redis cache as primary with in-memory fallback")
	}

	return &CacheManager{
		primary:     primary,
		fallback:    fallback,
		useFallback: useFallback,
	}
}

func (cm *CacheManager) Get(key string) (*CacheItem, bool) {
	item, found := cm.primary.Get(key)
	if found {
		return item, true
	}

	if !cm.useFallback {
		return nil, false
	}

	return cm.fallback.Get(key)
}

func (cm *CacheManager) Set(key string, data []byte, headers map[string]string, ttl time.Duration) error {

	if err := cm.primary.Set(key, data, headers, ttl); err != nil {
		if !cm.useFallback {
			log.Printf("Primary cache set failed, using fallback: %v", err)
			return cm.fallback.Set(key, data, headers, ttl)
		}
		return err
	}
	return nil
}

func (cm *CacheManager) Delete(key string) error {
	err1 := cm.primary.Delete(key)
	if !cm.useFallback {
		err2 := cm.fallback.Delete(key)
		if err1 != nil {
			return err1
		}
		return err2
	}
	return err1
}

func (cm *CacheManager) Clear() error {
	err1 := cm.primary.Clear()
	if !cm.useFallback {
		err2 := cm.fallback.Clear()
		if err1 != nil {
			return err1
		}
		return err2
	}
	return err1
}

func (cm *CacheManager) InvalidatePattern(pattern string) error {
	err1 := cm.primary.InvalidatePattern(pattern)
	if !cm.useFallback {
		err2 := cm.fallback.InvalidatePattern(pattern)
		if err1 != nil {
			return err1
		}
		return err2
	}
	return err1
}

func (cm *CacheManager) GetStats() CacheStats {
	return cm.primary.GetStats()
}

func (cm *CacheManager) ResetStats() error {
	return cm.primary.ResetStats()
}

var cacheManager *CacheManager

func InitializeCache() {
	cacheManager = NewCacheManager()
}

func generateRedisCacheKey(c *gin.Context) string {
	url := c.Request.URL.String()
	method := c.Request.Method
	userAgent := c.GetHeader("User-Agent")

	resource := ""
	if strings.Contains(url, "/posts") {
		resource = "posts"
	} else if strings.Contains(url, "/pages") {
		resource = "pages"
	} else if strings.Contains(url, "/media") {
		resource = "media"
	}

	key := fmt.Sprintf("%s:%s:%s", method, url, userAgent)
	hash := md5.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	if resource != "" {
		return resource + ":" + hashStr
	}
	return hashStr
}

func RedisCacheMiddleware(ttl time.Duration) gin.HandlerFunc {

	if cacheManager == nil {
		InitializeCache()
	}

	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		if strings.Contains(c.Request.URL.Path, "/admin") ||
			strings.Contains(c.Request.URL.Path, "/auth") {
			c.Next()
			return
		}

		cacheKey := generateRedisCacheKey(c)

		if cachedItem, exists := cacheManager.Get(cacheKey); exists {

			for key, value := range cachedItem.Headers {
				c.Header(key, value)
			}
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", cacheKey[:8]) // First 8 chars for debugging
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

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 && len(writer.body) > 0 {
			for key, values := range c.Writer.Header() {
				if len(values) > 0 {
					writer.headers[key] = values[0]
				}
			}

			if err := cacheManager.Set(cacheKey, writer.body, writer.headers, ttl); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}

			c.Header("X-Cache", "MISS")
			c.Header("X-Cache-Key", cacheKey[:8])
		}
	}
}

func GetCacheStats() CacheStats {
	if cacheManager == nil {
		return CacheStats{CacheType: "not_initialized"}
	}
	return cacheManager.GetStats()
}

func InvalidateCache(pattern string) error {
	if cacheManager == nil {
		return fmt.Errorf("cache manager not initialized")
	}
	return cacheManager.InvalidatePattern(pattern)
}

func ClearAllCache() error {
	if cacheManager == nil {
		return fmt.Errorf("cache manager not initialized")
	}
	return cacheManager.Clear()
}

func InvalidateMediaCache() error {
	return InvalidateCache("media")
}

func InvalidatePostCache() error {
	return InvalidateCache("posts")
}

func InvalidatePageCache() error {
	return InvalidateCache("pages")
}
