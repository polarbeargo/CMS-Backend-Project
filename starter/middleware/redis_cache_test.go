package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisCacheTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *RedisCacheTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("REDIS_PREFIX", "test_cache:")

	suite.router = gin.New()

	InitializeCache()
	suite.router.Use(RedisCacheMiddleware(1 * time.Minute))

	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "test response",
			"time":    time.Now().Unix(),
		})
	})

	suite.router.GET("/test-params", func(c *gin.Context) {
		name := c.Query("name")
		c.JSON(http.StatusOK, gin.H{
			"message": "hello " + name,
			"time":    time.Now().Unix(),
		})
	})
}

func (suite *RedisCacheTestSuite) TearDownSuite() {
	if cacheManager != nil {
		cacheManager.Clear()
	}
}

func (suite *RedisCacheTestSuite) TestCacheManagerInitialization() {
	assert.NotNil(suite.T(), cacheManager, "Cache manager should be initialized")

	stats := cacheManager.GetStats()
	assert.NotEmpty(suite.T(), stats.CacheType, "Cache type should be set")
	assert.True(suite.T(), stats.CacheType == "redis" || stats.CacheType == "in_memory",
		"Cache type should be redis or in_memory")
}

func (suite *RedisCacheTestSuite) TestCacheHitAndMiss() {
	cacheManager.Clear()
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	suite.router.ServeHTTP(w1, req1)

	assert.Equal(suite.T(), http.StatusOK, w1.Code)
	assert.Equal(suite.T(), "MISS", w1.Header().Get("X-Cache"))

	var response1 map[string]interface{}
	err := json.Unmarshal(w1.Body.Bytes(), &response1)
	assert.NoError(suite.T(), err)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)
	assert.Equal(suite.T(), "HIT", w2.Header().Get("X-Cache"))

	var response2 map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), response1["time"], response2["time"])
}

func (suite *RedisCacheTestSuite) TestCacheKeyGeneration() {
	cacheManager.Clear()
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test-params?name=john", nil)
	suite.router.ServeHTTP(w1, req1)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test-params?name=jane", nil)
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), "MISS", w1.Header().Get("X-Cache"))
	assert.Equal(suite.T(), "MISS", w2.Header().Get("X-Cache"))

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test-params?name=john", nil)
	suite.router.ServeHTTP(w3, req3)

	assert.Equal(suite.T(), "HIT", w3.Header().Get("X-Cache"))
}

func (suite *RedisCacheTestSuite) TestPOSTRequestsNotCached() {
	cacheManager.Clear()

	suite.router.POST("/test-post", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "post response"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test-post", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Empty(suite.T(), w.Header().Get("X-Cache"), "POST requests should not be cached")
}

func (suite *RedisCacheTestSuite) TestCacheInvalidation() {
	testKey := "test_key"
	testData := []byte("test data")
	testHeaders := map[string]string{"Content-Type": "application/json"}

	err := cacheManager.Set(testKey, testData, testHeaders, 1*time.Minute)
	assert.NoError(suite.T(), err)

	item, found := cacheManager.Get(testKey)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), testData, item.Data)

	err = cacheManager.InvalidatePattern("test")
	assert.NoError(suite.T(), err)

	_, found = cacheManager.Get(testKey)
	assert.False(suite.T(), found)
}

func (suite *RedisCacheTestSuite) TestCacheStats() {
	cacheManager.Clear()
	initialStats := cacheManager.GetStats()

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		suite.router.ServeHTTP(w, req)
	}

	finalStats := cacheManager.GetStats()
	assert.NotEmpty(suite.T(), finalStats.CacheType)

	hitsIncrease := finalStats.Hits - initialStats.Hits
	missesIncrease := finalStats.Misses - initialStats.Misses

	assert.True(suite.T(), hitsIncrease >= 4, "Should have at least 4 additional cache hits")
	assert.Equal(suite.T(), int64(1), missesIncrease, "Should have exactly 1 additional cache miss")
}

func (suite *RedisCacheTestSuite) TestCacheItemMethods() {
	testKey := "method_test_key"
	testData := []byte(`{"test": "data"}`)
	testHeaders := map[string]string{
		"Content-Type": "application/json",
		"X-Custom":     "header",
	}

	err := cacheManager.Set(testKey, testData, testHeaders, 1*time.Minute)
	assert.NoError(suite.T(), err)

	item, found := cacheManager.Get(testKey)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), testData, item.Data)
	assert.Equal(suite.T(), testHeaders, item.Headers)

	err = cacheManager.Delete(testKey)
	assert.NoError(suite.T(), err)

	_, found = cacheManager.Get(testKey)
	assert.False(suite.T(), found)
}

func (suite *RedisCacheTestSuite) TestCacheExpiration() {
	testKey := "expiration_test_key"
	testData := []byte("expiring data")
	testHeaders := map[string]string{"Content-Type": "text/plain"}

	err := cacheManager.Set(testKey, testData, testHeaders, 100*time.Millisecond)
	assert.NoError(suite.T(), err)

	_, found := cacheManager.Get(testKey)
	assert.True(suite.T(), found)

	time.Sleep(200 * time.Millisecond)

	_, found = cacheManager.Get(testKey)
	assert.False(suite.T(), found)
}

func (suite *RedisCacheTestSuite) TestInMemoryFallback() {
	// This test requires Redis to be unavailable
	// In a real test environment, you would temporarily disable Redis
	// For now, we'll test the in-memory cache directly

	inMemCache := NewInMemoryCache()

	testKey := "fallback_test"
	testData := []byte("fallback data")
	testHeaders := map[string]string{"Content-Type": "text/plain"}

	err := inMemCache.Set(testKey, testData, testHeaders, 1*time.Minute)
	assert.NoError(suite.T(), err)

	item, found := inMemCache.Get(testKey)
	assert.True(suite.T(), found)
	assert.Equal(suite.T(), testData, item.Data)

	err = inMemCache.Delete(testKey)
	assert.NoError(suite.T(), err)

	_, found = inMemCache.Get(testKey)
	assert.False(suite.T(), found)
}

func TestRedisCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}

func BenchmarkCacheSet(b *testing.B) {
	InitializeCache()

	testData := []byte("benchmark test data")
	testHeaders := map[string]string{"Content-Type": "application/json"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_" + string(rune(i))
		cacheManager.Set(key, testData, testHeaders, 1*time.Minute)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	InitializeCache()

	testData := []byte("benchmark test data")
	testHeaders := map[string]string{"Content-Type": "application/json"}

	for i := 0; i < 100; i++ {
		key := "benchmark_" + string(rune(i))
		cacheManager.Set(key, testData, testHeaders, 1*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "benchmark_" + string(rune(i%100))
		cacheManager.Get(key)
	}
}

func BenchmarkCacheMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	InitializeCache()

	router := gin.New()
	router.Use(RedisCacheMiddleware(1 * time.Minute))
	router.GET("/benchmark", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "benchmark response"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/benchmark", nil)
		router.ServeHTTP(w, req)
	}
}
