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
