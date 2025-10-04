package routes

import (
	"cms-backend/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitializeRoutes sets up all API routes
func InitializeRoutes(router *gin.Engine, db *gorm.DB) {

	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	api := router.Group("/api/v1")

	pages := api.Group("/pages")
	{
		pages.GET("", controllers.GetPages)
		pages.GET("/:id", controllers.GetPage)
		pages.POST("", controllers.CreatePage)
		pages.PUT("/:id", controllers.UpdatePage)
		pages.DELETE("/:id", controllers.DeletePage)
	}

	posts := api.Group("/posts")
	{
		posts.GET("", controllers.GetPosts)
		posts.GET("/:id", controllers.GetPost)
		posts.POST("", controllers.CreatePost)
		posts.PUT("/:id", controllers.UpdatePost)
		posts.DELETE("/:id", controllers.DeletePost)
	}

	media := api.Group("/media")
	{
		media.GET("", controllers.GetMedia)
		media.GET("/:id", controllers.GetMediaByID)
		media.POST("", controllers.CreateMedia)
		media.DELETE("/:id", controllers.DeleteMedia)
	}

	cache := api.Group("/cache")
	{
		cache.GET("/stats", controllers.GetCacheStats)
		cache.POST("/clear", controllers.ClearCache)
		cache.POST("/invalidate", controllers.InvalidateCache)
		cache.POST("/invalidate/:resource", controllers.InvalidateResourceCache)
		cache.POST("/warmup", controllers.WarmupCache)
		cache.GET("/health", controllers.CacheHealth)
	}
}
