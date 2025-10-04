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

var mediaValidator = validator.New()

func GetMedia(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var media []models.Media

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
	case "url", "type", "created_at", "updated_at":
	default:
		sortField = "created_at"
	}

	search := c.Query("search")
	mediaType := c.Query("type")

	query := db.Model(&models.Media{})
	var conditions []string
	var args []interface{}

	if search != "" {
		searchPattern := "%" + search + "%"
		conditions = append(conditions, "url ILIKE ? OR type ILIKE ?")
		args = append(args, searchPattern, searchPattern)
	}

	if mediaType != "" {
		conditions = append(conditions, "type = ?")
		args = append(args, mediaType)
	}

	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		query = query.Where(whereClause, args...)
	}

	var total int64
	countQuery := db.Model(&models.Media{})
	if len(conditions) > 0 {
		whereClause := strings.Join(conditions, " AND ")
		countQuery = countQuery.Where(whereClause, args...)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to count media records",
		})
		return
	}

	if err := query.Order(sortField + " " + sortOrder).
		Limit(pageSize).
		Offset(offset).
		Find(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to fetch media records",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       media,
		"page":       page,
		"page_size":  pageSize,
		"total":      total,
		"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func GetMediaByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid media ID"})
		return
	}
	var media models.Media
	if err := db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Media not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, media)
}

func CreateMedia(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var input models.Media
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: err.Error()})
		return
	}
	if err := mediaValidator.Struct(input); err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Validation failed: " + err.Error()})
		return
	}
	tx := db.Begin()
	if err := tx.Create(&input).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()

	middleware.InvalidateMediaCache()

	c.JSON(http.StatusCreated, input)
}

func DeleteMedia(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.HTTPError{Code: 400, Message: "Invalid media ID"})
		return
	}
	var media models.Media
	if err := db.First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, utils.HTTPError{Code: 404, Message: "Media not found"})
		} else {
			c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		}
		return
	}
	tx := db.Begin()
	if err := tx.Delete(&media).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	tx.Commit()

	middleware.InvalidateMediaCache()

	c.JSON(http.StatusOK, utils.MessageResponse{Message: "Media deleted"})
}
