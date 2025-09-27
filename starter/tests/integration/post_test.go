package integration

import (
	"cms-backend/models"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostIntegration(t *testing.T) {
	clearTables()
	mediaID := createTestMedia(t)

	t.Run("Create Post with Media", func(t *testing.T) {
		body := `{
            "title": "Test Post",
            "content": "Test Content",
            "author": "Tester",
            "media_ids": [` + fmt.Sprintf("%d", mediaID) + `]
        }`
		req := httptest.NewRequest("POST", "/api/v1/posts", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}
		var response models.Post
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.Title != "Test Post" || len(response.Media) == 0 {
			t.Errorf("Expected media attached to post")
		}
	})

	t.Run("Get Posts with Filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/posts?title=Test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		data, ok := response["data"].([]interface{})
		if !ok {
			t.Fatalf("Expected data array in response")
		}
		if len(data) == 0 {
			t.Errorf("Expected filtered post, got empty data")
		}
		firstPost := data[0].(map[string]interface{})
		if firstPost["title"] != "Test Post" {
			t.Errorf("Expected title 'Test Post', got %s", firstPost["title"])
		}
	})
}

func createTestMedia(t *testing.T) uint {
	body := `{
        "url": "http://example.com/test.jpg",
        "type": "image"
    }`
	req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test media, status: %d, body: %s", w.Code, w.Body.String())
	}
	var response models.Media
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to create test media: %v", err)
	}
	return response.ID
}

/*
TESTING HINTS:
1. Request Creation:
   - Use proper JSON formatting for relationships
   - Handle URL encoding for query parameters
   - Set appropriate headers

2. Response Validation:
   - Check both status codes and response content
   - Verify relationship data is correct
   - Validate filtered results carefully

3. Test Data:
   - Create meaningful test data
   - Handle relationships properly
   - Clean up between tests

4. Error Cases to Consider:
   - Invalid media IDs
   - Missing required fields
   - Invalid filter parameters
   - Non-existent relationships
*/
