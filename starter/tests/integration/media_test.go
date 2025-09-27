package integration

import (
	"cms-backend/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestMediaIntegration(t *testing.T) {
	clearTables()

	t.Run("Create Media", func(t *testing.T) {
		body := `{
            "url": "http://example.com/test.jpg",
            "type": "image"
        }`
		req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}
		var response models.Media
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.URL != "http://example.com/test.jpg" {
			t.Errorf("Expected URL 'http://example.com/test.jpg', got %s", response.URL)
		}
	})

	t.Run("Get All Media", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/media", nil)
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
			t.Errorf("Expected at least 1 media, got %d", len(data))
		}
	})

	t.Run("Get Media By ID", func(t *testing.T) {
		body := `{"url": "http://example.com/unique.jpg", "type": "image"}`
		req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var created models.Media
		json.Unmarshal(w.Body.Bytes(), &created)

		req = httptest.NewRequest("GET", "/api/v1/media/"+strconv.Itoa(int(created.ID)), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		var response models.Media
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}
		if response.URL != "http://example.com/unique.jpg" {
			t.Errorf("Expected URL 'http://example.com/unique.jpg', got %s", response.URL)
		}
	})

	t.Run("Delete Media", func(t *testing.T) {
		body := `{"url": "http://example.com/delete.jpg", "type": "image"}`
		req := httptest.NewRequest("POST", "/api/v1/media", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var created models.Media
		json.Unmarshal(w.Body.Bytes(), &created)

		req = httptest.NewRequest("DELETE", "/api/v1/media/"+strconv.Itoa(int(created.ID)), nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
	})
}

/*
TESTING HINTS:
1. Request Creation:
   - Use httptest.NewRequest for creating requests
   - Remember to set Content-Type for POST requests
   - Use strings.NewReader for request bodies

2. Response Handling:
   - Use httptest.NewRecorder for capturing responses
   - Parse JSON responses carefully
   - Check both status codes and response bodies

3. Test Data:
   - Use meaningful test data
   - Clean up between tests
   - Consider edge cases

4. Error Cases:
   - Test invalid inputs
   - Test missing required fields
   - Test invalid content types
*/
