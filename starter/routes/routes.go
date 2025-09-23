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

	api.GET("/pages", controllers.GetPages)
	api.GET("/pages/:id", controllers.GetPage)
	api.POST("/pages", controllers.CreatePage)
	api.PUT("/pages/:id", controllers.UpdatePage)
	api.DELETE("/pages/:id", controllers.DeletePage)

	api.GET("/posts", controllers.GetPosts)
	api.GET("/posts/:id", controllers.GetPost)
	api.POST("/posts", controllers.CreatePost)
	api.PUT("/posts/:id", controllers.UpdatePost)
	api.DELETE("/posts/:id", controllers.DeletePost)

	api.GET("/media", controllers.GetMedia)
	api.GET("/media/:id", controllers.GetMediaByID)
	api.POST("/media", controllers.CreateMedia)
	api.DELETE("/media/:id", controllers.DeleteMedia)
}
