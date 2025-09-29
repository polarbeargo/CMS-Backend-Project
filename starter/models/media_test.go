package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMediaModel(t *testing.T) {
	t.Run("MediaCreation", func(t *testing.T) {

		media := Media{
			ID:   1,
			URL:  "https://example.com/image.jpg",
			Type: "image/jpeg",
		}

		assert.Equal(t, uint(1), media.ID)
		assert.Equal(t, "https://example.com/image.jpg", media.URL)
		assert.Equal(t, "image/jpeg", media.Type)
	})

	t.Run("MediaValidation", func(t *testing.T) {

		media := Media{}

		assert.Empty(t, media.URL)
		assert.Empty(t, media.Type)
		assert.Equal(t, uint(0), media.ID)
	})

	t.Run("MediaTimestamps", func(t *testing.T) {
		now := time.Now()
		media := Media{
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.False(t, media.CreatedAt.IsZero())
		assert.False(t, media.UpdatedAt.IsZero())
		assert.True(t, media.CreatedAt.Equal(now))
		assert.True(t, media.UpdatedAt.Equal(now))
	})

	t.Run("MediaTypes", func(t *testing.T) {
		testCases := []struct {
			name      string
			url       string
			mediaType string
		}{
			{"Image", "https://example.com/photo.jpg", "image/jpeg"},
			{"Video", "https://example.com/video.mp4", "video/mp4"},
			{"Audio", "https://example.com/audio.mp3", "audio/mp3"},
			{"Document", "https://example.com/doc.pdf", "application/pdf"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				media := Media{
					URL:  tc.url,
					Type: tc.mediaType,
				}

				assert.Equal(t, tc.url, media.URL)
				assert.Equal(t, tc.mediaType, media.Type)
				assert.Contains(t, media.URL, "https://")
			})
		}
	})

	t.Run("MediaURLValidation", func(t *testing.T) {
		validURLs := []string{
			"https://example.com/image.jpg",
			"http://localhost:8080/media/file.png",
			"https://cdn.example.com/assets/video.mp4",
		}

		for _, url := range validURLs {
			media := Media{
				URL:  url,
				Type: "test/type",
			}

			assert.NotEmpty(t, media.URL)
			assert.True(t, len(media.URL) > 0)
		}
	})
}
