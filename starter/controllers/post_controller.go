package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var postValidator = validator.New()

type PostInput struct {
	Title    string `json:"title" validate:"required"`
	Content  string `json:"content" validate:"required"`
	Author   string `json:"author"`
	MediaIDs []uint `json:"media_ids"`
}

func GetPosts(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var posts []models.Post

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := strings.ToLower(c.DefaultQuery("sort_order", "desc"))
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}
	sortField := sortBy
	switch sortBy {
	case "title", "author", "created_at", "updated_at":
	default:
		sortField = "created_at"
	}

	search := c.Query("search")
	query := db.Model(&models.Post{}).Preload("Media")
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			db.Where("title ILIKE ?", searchPattern).
				Or("content ILIKE ?", searchPattern).
				Or("author ILIKE ?", searchPattern),
		)
	}

	title := c.Query("title")
	author := c.Query("author")
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author = ?", author)
	}

	var total int64
	query.Count(&total)

	if err := query.Order(sortField + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       posts,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
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
	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if err := postValidator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Validation failed: " + err.Error()})
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
	var input PostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if err := postValidator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Validation failed: " + err.Error()})
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
