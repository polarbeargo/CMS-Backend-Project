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

func TestGetPages(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "pages"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "First Page", "Content 1", time.Now(), time.Now()).
		AddRow(2, "Second Page", "Content 2", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "pages"`).WillReturnRows(rows)

	router.GET("/pages", GetPages)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response gin.H
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	data := response["data"].([]interface{})
	if len(data) != 2 {
		t.Fatalf("Expected 2 pages, but got %d", len(data))
	}
}

func TestGetPage(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "First Page", "Content 1", now, now)

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	router.GET("/pages/:id", GetPage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.ID != 1 || response.Title != "First Page" {
		t.Fatalf("Unexpected page data: %+v", response)
	}
}

func TestCreatePage(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "pages"`).
		WithArgs("New Page", "New Content", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	page := models.Page{Title: "New Page", Content: "New Content"}
	body, _ := json.Marshal(page)

	router.POST("/pages", CreatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pages", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Title != "New Page" {
		t.Fatalf("Expected title 'New Page', got %s", response.Title)
	}
}

func TestUpdatePage(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Old Title", "Old Content", now, now)
	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "pages" SET "title"=\$1,"content"=\$2,"created_at"=\$3,"updated_at"=\$4 WHERE "id" = \$5`).
		WithArgs("Updated Title", "Updated Content", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	update := models.Page{Title: "Updated Title", Content: "Updated Content"}
	body, _ := json.Marshal(update)

	router.PUT("/pages/:id", UpdatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/pages/1", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response models.Page
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Title != "Updated Title" {
		t.Fatalf("Expected updated title, got %s", response.Title)
	}
}

func TestDeletePage(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Title", "Content", now, now)
	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "pages" WHERE "pages"\."id" = \$1`).WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	router.DELETE("/pages/:id", DeletePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/pages/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response utils.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Page deleted" {
		t.Fatalf("Expected deletion message, got %s", response.Message)
	}
}

func TestGetPage_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.GET("/pages/:id", GetPage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid page ID" {
		t.Fatalf("Expected 'Invalid page ID', got %s", response.Message)
	}
}

func TestGetPage_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.GET("/pages/:id", GetPage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Page not found" {
		t.Fatalf("Expected 'Page not found', got %s", response.Message)
	}
}

func TestGetPage_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/pages/:id", GetPage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pages/1", nil)
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

func TestCreatePage_InvalidJSON(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.POST("/pages", CreatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pages", strings.NewReader("{invalid json}"))
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

func TestCreatePage_MissingRequiredFields(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	incompletePage := map[string]string{
		"title":   "",
		"content": "Some content",
	}
	body, _ := json.Marshal(incompletePage)

	router.POST("/pages", CreatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pages", strings.NewReader(string(body)))
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

func TestCreatePage_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "pages"`).
		WithArgs("Test Page", "Test Content", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	page := models.Page{Title: "Test Page", Content: "Test Content"}
	body, _ := json.Marshal(page)

	router.POST("/pages", CreatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pages", strings.NewReader(string(body)))
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

func TestUpdatePage_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	updateData := models.Page{Title: "Updated", Content: "Updated Content"}
	body, _ := json.Marshal(updateData)

	router.PUT("/pages/:id", UpdatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/pages/invalid", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid page ID" {
		t.Fatalf("Expected 'Invalid page ID', got %s", response.Message)
	}
}

func TestUpdatePage_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	updateData := models.Page{Title: "Updated", Content: "Updated Content"}
	body, _ := json.Marshal(updateData)

	router.PUT("/pages/:id", UpdatePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/pages/999", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Page not found" {
		t.Fatalf("Expected 'Page not found', got %s", response.Message)
	}
}

func TestDeletePage_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.DELETE("/pages/:id", DeletePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/pages/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid page ID" {
		t.Fatalf("Expected 'Invalid page ID', got %s", response.Message)
	}
}

func TestDeletePage_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.DELETE("/pages/:id", DeletePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/pages/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Page not found" {
		t.Fatalf("Expected 'Page not found', got %s", response.Message)
	}
}

func TestDeletePage_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"}).
		AddRow(1, "Test Page", "Test Content", now, now)
	mock.ExpectQuery(`SELECT \* FROM "pages" WHERE "pages"\."id" = \$1 ORDER BY "pages"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "pages" WHERE "pages"\."id" = \$1`).
		WithArgs(1).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	router.DELETE("/pages/:id", DeletePage)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/pages/1", nil)
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

func TestGetPages_WithPaginationAndSorting(t *testing.T) {

	testCases := []struct {
		name         string
		queryParams  string
		expectStatus int
	}{
		{"ValidPagination", "?page=2&page_size=5", http.StatusOK},
		{"InvalidPage", "?page=0", http.StatusOK},           // 應該預設為 1
		{"InvalidPageSize", "?page_size=-1", http.StatusOK}, // 應該預設為 10
		{"ValidSorting", "?sort_by=title&sort_order=asc", http.StatusOK},
		{"InvalidSorting", "?sort_by=invalid&sort_order=invalid", http.StatusOK},
		{"SearchFilter", "?search=test", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, _, mock := utils.SetupRouterAndMockDB(t)
			defer mock.ExpectClose()

			countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)

			if tc.name == "SearchFilter" {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "pages" WHERE title ILIKE \$1 OR content ILIKE \$2`).
					WithArgs("%test%", "%test%").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "pages" WHERE title ILIKE \$1 OR content ILIKE \$2 ORDER BY created_at desc LIMIT \$3`).
					WithArgs("%test%", "%test%", 10).
					WillReturnRows(rows)
			} else {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "pages"`).WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "title", "content", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "pages"`).WillReturnRows(rows)
			}

			router.GET("/pages", GetPages)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/pages"+tc.queryParams, nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.expectStatus {
				t.Fatalf("Expected status %d for %s, got %d", tc.expectStatus, tc.name, w.Code)
			}
		})
	}
}

/*
TESTING HINTS:
1. Use sqlmock.AnyArg() for timestamp fields
2. Remember to escape special characters in SQL patterns
3. Each database operation needs proper error handling
4. Content-Type header is required for POST/PUT requests
5. Transaction tests need Begin/Commit expectations
6. Use proper argument matching in mock expectations
7. Consider testing error cases:
   - Invalid IDs
   - Missing required fields
   - Database errors
*/
