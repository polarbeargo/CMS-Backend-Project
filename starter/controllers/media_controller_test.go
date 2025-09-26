package controllers

import (
	"cms-backend/models"
	"cms-backend/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestGetMedia(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "media"`).WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image1.jpg", "image", time.Now(), time.Now()).
		AddRow(2, "http://example.com/image2.jpg", "image", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "media" ORDER BY created_at desc LIMIT \$1`).
		WithArgs(10).
		WillReturnRows(rows)

	router.GET("/media", GetMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}

	data, ok := response["data"].([]interface{})
	if !ok {
		t.Fatalf("Expected data array in response")
	}
	if len(data) != 2 {
		t.Fatalf("Expected 2 media, but got %d", len(data))
	}
}

func TestGetMediaByID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image1.jpg", "image", now, now)

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	router.GET("/media/:id", GetMediaByID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.ID != 1 || response.URL != "http://example.com/image1.jpg" {
		t.Fatalf("Unexpected media data: %+v", response)
	}
}

func TestCreateMedia(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "media"`).
		WithArgs("http://example.com/image3.jpg", "image", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	media := models.Media{URL: "http://example.com/image3.jpg", Type: "image"}
	body, _ := json.Marshal(media)

	router.POST("/media", CreateMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/media", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}

	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.URL != "http://example.com/image3.jpg" {
		t.Fatalf("Expected URL 'http://example.com/image3.jpg', got %s", response.URL)
	}
}

func TestDeleteMedia(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image1.jpg", "image", now, now)
	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "media" WHERE "media"\."id" = \$1`).WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	router.DELETE("/media/:id", DeleteMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/media/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response utils.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Media deleted" {
		t.Fatalf("Expected deletion message, got %s", response.Message)
	}
}
