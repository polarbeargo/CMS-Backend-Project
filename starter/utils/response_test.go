package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageResponse(t *testing.T) {
	t.Run("MessageResponseCreation", func(t *testing.T) {

		response := MessageResponse{
			Message: "Operation successful",
		}

		assert.Equal(t, "Operation successful", response.Message)
		assert.NotEmpty(t, response.Message)
	})

	t.Run("MessageResponseJSON", func(t *testing.T) {

		response := MessageResponse{
			Message: "Data created successfully",
		}

		assert.Equal(t, "Data created successfully", response.Message)

		testMessages := []string{
			"Media deleted",
			"Page created successfully",
			"Post updated",
			"User authenticated",
			"File uploaded",
		}

		for _, msg := range testMessages {
			resp := MessageResponse{Message: msg}
			assert.Equal(t, msg, resp.Message)
			assert.NotEmpty(t, resp.Message)
		}
	})
}

func TestHTTPError(t *testing.T) {
	t.Run("HTTPErrorCreation", func(t *testing.T) {
		httpError := HTTPError{
			Code:    400,
			Message: "Bad Request",
		}

		assert.Equal(t, 400, httpError.Code)
		assert.Equal(t, "Bad Request", httpError.Message)
	})

	t.Run("HTTPErrorValidation", func(t *testing.T) {
		emptyError := HTTPError{}

		assert.Equal(t, 0, emptyError.Code)
		assert.Empty(t, emptyError.Message)
	})

	t.Run("CommonHTTPErrors", func(t *testing.T) {
		testCases := []struct {
			name    string
			code    int
			message string
		}{
			{"BadRequest", 400, "Bad Request"},
			{"Unauthorized", 401, "Unauthorized"},
			{"Forbidden", 403, "Forbidden"},
			{"NotFound", 404, "Not Found"},
			{"MethodNotAllowed", 405, "Method Not Allowed"},
			{"Conflict", 409, "Conflict"},
			{"UnprocessableEntity", 422, "Unprocessable Entity"},
			{"InternalServerError", 500, "Internal Server Error"},
			{"BadGateway", 502, "Bad Gateway"},
			{"ServiceUnavailable", 503, "Service Unavailable"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				httpError := HTTPError{
					Code:    tc.code,
					Message: tc.message,
				}

				assert.Equal(t, tc.code, httpError.Code)
				assert.Equal(t, tc.message, httpError.Message)
				assert.True(t, httpError.Code >= 400 && httpError.Code < 600,
					"HTTP error code should be in 4xx or 5xx range")
			})
		}
	})

	t.Run("HTTPErrorTypes", func(t *testing.T) {

		clientError := HTTPError{Code: 400, Message: "Invalid input"}
		serverError := HTTPError{Code: 500, Message: "Database connection failed"}

		assert.True(t, clientError.Code >= 400 && clientError.Code < 500,
			"Should be client error")
		assert.Equal(t, "Invalid input", clientError.Message)

		assert.True(t, serverError.Code >= 500 && serverError.Code < 600,
			"Should be server error")
		assert.Equal(t, "Database connection failed", serverError.Message)
	})

	t.Run("HTTPErrorValidation", func(t *testing.T) {
		validError := HTTPError{Code: 404, Message: "Resource not found"}
		invalidCodeError := HTTPError{Code: 200, Message: "This should be an error"}

		assert.True(t, validError.Code >= 400, "Error code should be >= 400")
		assert.NotEmpty(t, validError.Message, "Error message should not be empty")

		assert.False(t, invalidCodeError.Code >= 400,
			"2xx codes should not be used for errors")
	})
}
