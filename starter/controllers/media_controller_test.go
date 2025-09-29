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
	"gorm.io/gorm"
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

func TestGetMediaByID_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.GET("/media/:id", GetMediaByID)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid media ID" {
		t.Fatalf("Expected 'Invalid media ID', got %s", response.Message)
	}
}

func TestGetMediaByID_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.GET("/media/:id", GetMediaByID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Media not found" {
		t.Fatalf("Expected 'Media not found', got %s", response.Message)
	}
}

func TestGetMediaByID_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/media/:id", GetMediaByID)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/media/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Code != 500 {
		t.Fatalf("Expected error code 500, got %d", response.Code)
	}
}

func TestCreateMedia_InvalidJSON(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.POST("/media", CreateMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/media", strings.NewReader("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Code != 400 {
		t.Fatalf("Expected error code 400, got %d", response.Code)
	}
}

func TestCreateMedia_MissingRequiredFields(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	incompleteMedia := map[string]string{
		"url":  "",
		"type": "image",
	}
	body, _ := json.Marshal(incompleteMedia)

	router.POST("/media", CreateMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/media", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Code != 400 {
		t.Fatalf("Expected error code 400, got %d", response.Code)
	}
}

func TestCreateMedia_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "media"`).
		WithArgs("http://example.com/image.jpg", "image", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	media := models.Media{URL: "http://example.com/image.jpg", Type: "image"}
	body, _ := json.Marshal(media)

	router.POST("/media", CreateMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/media", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Code != 500 {
		t.Fatalf("Expected error code 500, got %d", response.Code)
	}
}

func TestDeleteMedia_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.DELETE("/media/:id", DeleteMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/media/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid media ID" {
		t.Fatalf("Expected 'Invalid media ID', got %s", response.Message)
	}
}

func TestDeleteMedia_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.DELETE("/media/:id", DeleteMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/media/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Media not found" {
		t.Fatalf("Expected 'Media not found', got %s", response.Message)
	}
}

func TestDeleteMedia_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image.jpg", "image", now, now)
	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" = \$1 ORDER BY "media"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "media" WHERE "media"\."id" = \$1`).
		WithArgs(1).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	router.DELETE("/media/:id", DeleteMedia)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/media/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("Expected status 500, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Code != 500 {
		t.Fatalf("Expected error code 500, got %d", response.Code)
	}
}

func TestGetMedia_WithPaginationBoundaries(t *testing.T) {

	testCases := []struct {
		name             string
		queryParams      string
		expectedPage     int
		expectedPageSize int
	}{
		{"NegativePage", "?page=-1&page_size=10", 1, 10},
		{"ZeroPage", "?page=0&page_size=10", 1, 10},
		{"NegativePageSize", "?page=1&page_size=-5", 1, 10},
		{"ExcessivePageSize", "?page=1&page_size=200", 1, 10},
		{"DefaultValues", "", 1, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, _, mock := utils.SetupRouterAndMockDB(t)
			defer mock.ExpectClose()

			countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
			mock.ExpectQuery(`SELECT count\(\*\) FROM "media"`).WillReturnRows(countRows)

			rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"})
			mock.ExpectQuery(`SELECT \* FROM "media" ORDER BY created_at desc LIMIT \$1`).
				WithArgs(tc.expectedPageSize).
				WillReturnRows(rows)

			router.GET("/media", GetMedia)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/media"+tc.queryParams, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200, got %d", w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Error unmarshaling response: %v", err)
			}

			if int(response["page"].(float64)) != tc.expectedPage {
				t.Fatalf("Expected page %d, got %v", tc.expectedPage, response["page"])
			}
			if int(response["page_size"].(float64)) != tc.expectedPageSize {
				t.Fatalf("Expected page_size %d, got %v", tc.expectedPageSize, response["page_size"])
			}
		})
	}
}

func TestGetMedia_WithSortingAndFiltering(t *testing.T) {

	testCases := []struct {
		name        string
		queryParams string
		expectQuery string
	}{
		{"InvalidSortBy", "?sort_by=invalid_field", "created_at desc"},
		{"ValidSortBy", "?sort_by=url", "url desc"},
		{"InvalidSortOrder", "?sort_order=invalid", "created_at desc"},
		{"ValidSortOrder", "?sort_order=asc", "created_at asc"},
		{"SearchFilter", "?search=test", ""},
		{"TypeFilter", "?type=image", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, _, mock := utils.SetupRouterAndMockDB(t)
			defer mock.ExpectClose()

			countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)

			if tc.name == "SearchFilter" {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "media" WHERE url ILIKE \$1 OR type ILIKE \$2`).
					WithArgs("%test%", "%test%").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "media" WHERE url ILIKE \$1 OR type ILIKE \$2 ORDER BY created_at desc LIMIT \$3`).
					WithArgs("%test%", "%test%", 10).
					WillReturnRows(rows)
			} else if tc.name == "TypeFilter" {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "media" WHERE type = \$1`).
					WithArgs("image").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "media" WHERE type = \$1 ORDER BY created_at desc LIMIT \$2`).
					WithArgs("image", 10).
					WillReturnRows(rows)
			} else {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "media"`).WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"})
				if tc.expectQuery != "" {
					mock.ExpectQuery(`SELECT \* FROM "media" ORDER BY ` + tc.expectQuery + ` LIMIT \$1`).
						WithArgs(10).
						WillReturnRows(rows)
				} else {
					mock.ExpectQuery(`SELECT \* FROM "media" ORDER BY created_at desc LIMIT \$1`).
						WithArgs(10).
						WillReturnRows(rows)
				}
			}

			router.GET("/media", GetMedia)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/media"+tc.queryParams, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200 for %s, got %d", tc.name, w.Code)
			}
		})
	}
}
