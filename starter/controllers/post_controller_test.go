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

func TestGetPosts(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "posts"`).WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "First Post", "Content 1", "Author1", time.Now(), time.Now()).
		AddRow(2, "Second Post", "Content 2", "Author2", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "posts" ORDER BY created_at desc LIMIT \$1`).
		WithArgs(10).
		WillReturnRows(rows)

	postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" IN \(\$1,\$2\)`).
		WithArgs(1, 2).
		WillReturnRows(postMediaRows)

	router.GET("/posts", GetPosts)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
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
		t.Fatalf("Expected 2 posts, but got %d", len(data))
	}
}

func TestGetPostsWithFilters(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	countRows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "posts" WHERE title ILIKE \$1 AND author = \$2`).
		WithArgs("%Filtered%", "AuthorX").
		WillReturnRows(countRows)

	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Filtered Post", "Content", "AuthorX", time.Now(), time.Now())

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE title ILIKE \$1 AND author = \$2 ORDER BY created_at desc LIMIT \$3`).
		WithArgs("%Filtered%", "AuthorX", 10).
		WillReturnRows(rows)

	postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(1).
		WillReturnRows(postMediaRows)

	router.GET("/posts", GetPosts)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts?title=Filtered&author=AuthorX", nil)
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
	if len(data) != 1 {
		t.Fatalf("Expected 1 post, got %d", len(data))
	}

	post := data[0].(map[string]interface{})
	if post["title"] != "Filtered Post" {
		t.Fatalf("Expected filtered post title, got %s", post["title"])
	}
}

func TestGetPost(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "First Post", "Content 1", "Author1", now, now)

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(1).
		WillReturnRows(postMediaRows)

	router.GET("/posts/:id", GetPost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, but got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.ID != 1 || response.Title != "First Post" {
		t.Fatalf("Unexpected post data: %+v", response)
	}
}

func TestCreatePost(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "posts"`).
		WithArgs("New Post", "New Content", "Author", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	postRows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(3, "New Post", "New Content", "Author", time.Now(), time.Now())
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(3, 1).
		WillReturnRows(postRows)

	postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(3).
		WillReturnRows(postMediaRows)

	input := map[string]interface{}{
		"title":     "New Post",
		"content":   "New Content",
		"author":    "Author",
		"media_ids": []uint{},
	}
	body, _ := json.Marshal(input)

	router.POST("/posts", CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Title != "New Post" {
		t.Fatalf("Expected title 'New Post', got %s", response.Title)
	}
}

func TestUpdatePost(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Old Title", "Old Content", "Author", now, now)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
	mock.ExpectQuery(`SELECT \* FROM "post_media" WHERE "post_media"\."post_id" = \$1`).
		WithArgs(1).
		WillReturnRows(postMediaRows)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "posts" SET "title"=\$1,"content"=\$2,"author"=\$3,"created_at"=\$4,"updated_at"=\$5 WHERE "id" = \$6`).
		WithArgs("Updated Title", "Updated Content", "Author", sqlmock.AnyArg(), sqlmock.AnyArg(), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	update := map[string]interface{}{
		"title":   "Updated Title",
		"content": "Updated Content",
		"author":  "Author",
	}
	body, _ := json.Marshal(update)

	router.PUT("/posts/:id", UpdatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/posts/1", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response models.Post
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Title != "Updated Title" {
		t.Fatalf("Expected updated title, got %s", response.Title)
	}
}

func TestDeletePost(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Title", "Content", "Author", now, now)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).WithArgs(1, 1).WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "posts" WHERE "posts"\."id" = \$1`).WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	router.DELETE("/posts/:id", DeletePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response utils.MessageResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Post deleted" {
		t.Fatalf("Expected deletion message, got %s", response.Message)
	}
}

func TestGetPost_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.GET("/posts/:id", GetPost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid post ID" {
		t.Fatalf("Expected 'Invalid post ID', got %s", response.Message)
	}
}

func TestGetPost_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.GET("/posts/:id", GetPost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Post not found" {
		t.Fatalf("Expected 'Post not found', got %s", response.Message)
	}
}

func TestGetPost_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/posts/:id", GetPost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/1", nil)
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

func TestCreatePost_InvalidJSON(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.POST("/posts", CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader("{invalid json}"))
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

func TestCreatePost_MissingRequiredFields(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	incompletePost := map[string]interface{}{
		"title":   "",
		"content": "Some content",
		"author":  "Author",
	}
	body, _ := json.Marshal(incompletePost)

	router.POST("/posts", CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(string(body)))
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

func TestCreatePost_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "posts"`).
		WithArgs("Test Post", "Test Content", "Author", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	input := map[string]interface{}{
		"title":     "Test Post",
		"content":   "Test Content",
		"author":    "Author",
		"media_ids": []uint{},
	}
	body, _ := json.Marshal(input)

	router.POST("/posts", CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(string(body)))
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

func TestUpdatePost_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	updateData := map[string]string{
		"title":   "Updated",
		"content": "Updated Content",
	}
	body, _ := json.Marshal(updateData)

	router.PUT("/posts/:id", UpdatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/posts/invalid", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid post ID" {
		t.Fatalf("Expected 'Invalid post ID', got %s", response.Message)
	}
}

func TestUpdatePost_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	updateData := map[string]string{
		"title":   "Updated",
		"content": "Updated Content",
	}
	body, _ := json.Marshal(updateData)

	router.PUT("/posts/:id", UpdatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/posts/999", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Post not found" {
		t.Fatalf("Expected 'Post not found', got %s", response.Message)
	}
}

func TestDeletePost_InvalidID(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	router.DELETE("/posts/:id", DeletePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/posts/invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Invalid post ID" {
		t.Fatalf("Expected 'Invalid post ID', got %s", response.Message)
	}
}

func TestDeletePost_NotFound(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(999, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	router.DELETE("/posts/:id", DeletePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/posts/999", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d", w.Code)
	}

	var response utils.HTTPError
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Error unmarshaling response: %v", err)
	}
	if response.Message != "Post not found" {
		t.Fatalf("Expected 'Post not found', got %s", response.Message)
	}
}

func TestDeletePost_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"}).
		AddRow(1, "Test Post", "Test Content", "Author", now, now)
	mock.ExpectQuery(`SELECT \* FROM "posts" WHERE "posts"\."id" = \$1 ORDER BY "posts"\."id" LIMIT \$2`).
		WithArgs(1, 1).
		WillReturnRows(rows)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "posts" WHERE "posts"\."id" = \$1`).
		WithArgs(1).
		WillReturnError(gorm.ErrInvalidDB)
	mock.ExpectRollback()

	router.DELETE("/posts/:id", DeletePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/posts/1", nil)
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

func TestCreatePost_WithInvalidMediaIDs(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	input := map[string]interface{}{
		"title":     "Post with Media",
		"content":   "Content with media",
		"author":    "Author",
		"media_ids": []uint{999, 1000},
	}
	body, _ := json.Marshal(input)

	mock.ExpectQuery(`SELECT \* FROM "media" WHERE "media"\."id" IN \(\$1,\$2\)`).
		WithArgs(999, 1000).
		WillReturnRows(sqlmock.NewRows([]string{"id", "url", "type", "created_at", "updated_at"}))

	router.POST("/posts", CreatePost)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts", strings.NewReader(string(body)))
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

func TestGetPosts_DatabaseError(t *testing.T) {
	router, _, mock := utils.SetupRouterAndMockDB(t)
	defer mock.ExpectClose()

	mock.ExpectQuery(`SELECT count\(\*\) FROM "posts"`).
		WillReturnError(gorm.ErrInvalidDB)

	router.GET("/posts", GetPosts)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts", nil)
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

func TestGetPosts_WithComplexFilters(t *testing.T) {
	testCases := []struct {
		name         string
		queryParams  string
		expectStatus int
	}{
		{"EmptySearch", "?search=", http.StatusOK},
		{"EmptyAuthor", "?author=", http.StatusOK},
		{"CombinedFilters", "?search=test&author=TestAuthor&page=1&page_size=5", http.StatusOK},
		{"InvalidPagination", "?page=-1&page_size=0", http.StatusOK}, // 應該使用預設值
		{"LargePage", "?page=1000", http.StatusOK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, _, mock := utils.SetupRouterAndMockDB(t)
			defer mock.ExpectClose()

			countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)

			if tc.name == "CombinedFilters" {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "posts" WHERE \(title ILIKE \$1 OR content ILIKE \$2 OR author ILIKE \$3\) AND author = \$4`).
					WithArgs("%test%", "%test%", "%test%", "TestAuthor").
					WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "posts" WHERE \(title ILIKE \$1 OR content ILIKE \$2 OR author ILIKE \$3\) AND author = \$4 ORDER BY created_at desc LIMIT \$5`).
					WithArgs("%test%", "%test%", "%test%", "TestAuthor", 5).
					WillReturnRows(rows)

				postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
				mock.ExpectQuery(`SELECT \* FROM "post_media"`).WillReturnRows(postMediaRows)
			} else if tc.name == "LargePage" {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "posts"`).WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "posts" ORDER BY created_at desc LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 9990).
					WillReturnRows(rows)

				postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
				mock.ExpectQuery(`SELECT \* FROM "post_media"`).WillReturnRows(postMediaRows)
			} else {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "posts"`).WillReturnRows(countRows)

				rows := sqlmock.NewRows([]string{"id", "title", "content", "author", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT \* FROM "posts" ORDER BY created_at desc LIMIT \$1`).
					WithArgs(10).
					WillReturnRows(rows)

				postMediaRows := sqlmock.NewRows([]string{"post_id", "media_id"})
				mock.ExpectQuery(`SELECT \* FROM "post_media"`).WillReturnRows(postMediaRows)
			}

			router.GET("/posts", GetPosts)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/posts"+tc.queryParams, nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.expectStatus {
				t.Fatalf("Expected status %d for %s, got %d", tc.expectStatus, tc.name, w.Code)
			}
		})
	}
}
