package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetPosts retrieves all posts with optional filtering
func GetPosts(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var posts []models.Post

	title := c.Query("title")
	author := c.Query("author")

	query := db
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author = ?", author)
	}

	// Use proper preloading for media relationships
	if err := query.Preload("Media").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// GetPost retrieves a specific post by ID
func GetPost(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid post ID"})
		return
	}
	var post models.Post
	if err := db.Preload("Media").First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, post)
}

// CreatePost creates a new post
func CreatePost(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input struct {
		Title    string `json:"title" binding:"required"`
		Content  string `json:"content" binding:"required"`
		Author   string `json:"author"`
		MediaIDs []uint `json:"media_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	var media []models.Media
	if len(input.MediaIDs) > 0 {
		if err := db.Find(&media, input.MediaIDs).Error; err != nil {
			c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid media IDs"})
			return
		}
	}
	post := models.Post{
		Title:   input.Title,
		Content: input.Content,
		Author:  input.Author,
		Media:   media,
	}
	tx := db.Begin()
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()
	if err := db.Preload("Media").First(&post, post.ID).Error; err == nil {
		c.JSON(http.StatusCreated, post)
	} else {
		c.JSON(http.StatusCreated, post)
	}
}

// UpdatePost updates an existing post
func UpdatePost(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid post ID"})
		return
	}
	var post models.Post
	if err := db.Preload("Media").First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	var input struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Author   string `json:"author"`
		MediaIDs []uint `json:"media_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if input.Title != "" {
		post.Title = input.Title
	}
	if input.Content != "" {
		post.Content = input.Content
	}
	if input.Author != "" {
		post.Author = input.Author
	}
	if input.MediaIDs != nil {
		var media []models.Media
		if len(input.MediaIDs) > 0 {
			if err := db.Find(&media, input.MediaIDs).Error; err != nil {
				c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid media IDs"})
				return
			}
		}
		post.Media = media
	}
	tx := db.Begin()
	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()
	if err := db.Preload("Media").First(&post, post.ID).Error; err == nil {
		c.JSON(http.StatusOK, post)
	} else {
		c.JSON(http.StatusOK, post)
	}
}

// DeletePost deletes a post
func DeletePost(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid post ID"})
		return
	}
	var post models.Post
	if err := db.First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	tx := db.Begin()
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, utils.MessageResponse{Message: "Post deleted"})
}
