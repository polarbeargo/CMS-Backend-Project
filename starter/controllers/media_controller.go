package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var mediaValidator = validator.New()

func GetMedia(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	var media []models.Media
	if err := db.Find(&media).Error; err != nil {
		c.JSON(http.StatusInternalServerError, utils.HTTPError{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, media)
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
	c.JSON(http.StatusOK, utils.MessageResponse{Message: "Media deleted"})
}
