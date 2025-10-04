package controllers

import (
	"cms-backend/middleware"
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var pageValidator = validator.New()

func GetPages(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var pages []models.Page

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
	case "title", "created_at", "updated_at":
	default:
		sortField = "created_at"
	}

	search := c.Query("search")
	query := db.Model(&models.Page{})
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			db.Where("title ILIKE ?", searchPattern).
				Or("content ILIKE ?", searchPattern),
		)
	}

	title := c.Query("title")
	if title != "" {
		query = query.Where("title ILIKE ?", "%"+title+"%")
	}

	var total int64
	query.Count(&total)

	if err := query.Order(sortField + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&pages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       pages,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

// GetPage retrieves a specific page by ID
func GetPage(c *gin.Context) {
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: "Database connection error"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid page ID"})
		return
	}
	var page models.Page
	if err := db.First(&page, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, page)
}

// CreatePage creates a new page
func CreatePage(c *gin.Context) {
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: "Database connection error"})
		return
	}
	var page models.Page
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if err := pageValidator.Struct(page); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Validation failed: " + err.Error()})
		return
	}
	tx := db.Begin()
	if err := tx.Create(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()
	middleware.InvalidatePageCache()

	c.JSON(http.StatusCreated, page)
}

// UpdatePage updates an existing page by ID
func UpdatePage(c *gin.Context) {
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: "Database connection error"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid page ID"})
		return
	}
	var page models.Page
	if err := db.First(&page, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	var input models.Page
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if err := pageValidator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Validation failed: " + err.Error()})
		return
	}
	page.Title = input.Title
	page.Content = input.Content
	tx := db.Begin()
	if err := tx.Save(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()
	middleware.InvalidatePageCache()

	c.JSON(http.StatusOK, page)
}

// DeletePage deletes a page by ID
func DeletePage(c *gin.Context) {
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: "Database connection error"})
		return
	}
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid page ID"})
		return
	}
	var page models.Page
	if err := db.First(&page, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	tx := db.Begin()
	if err := tx.Delete(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()

	middleware.InvalidatePageCache()

	c.JSON(http.StatusOK, utils.MessageResponse{Message: "Page deleted"})
}
