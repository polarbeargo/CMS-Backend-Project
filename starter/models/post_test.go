package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPostModel(t *testing.T) {
	t.Run("PostCreation", func(t *testing.T) {
		post := Post{
			ID:      1,
			Title:   "Test Post",
			Content: "This is a test post content",
			Author:  "Test Author",
		}

		assert.Equal(t, uint(1), post.ID)
		assert.Equal(t, "Test Post", post.Title)
		assert.Equal(t, "This is a test post content", post.Content)
		assert.Equal(t, "Test Author", post.Author)
	})

	t.Run("PostValidation", func(t *testing.T) {
		post := Post{}

		assert.Empty(t, post.Title)
		assert.Empty(t, post.Content)
	})

	t.Run("PostTimestamps", func(t *testing.T) {
		now := time.Now()
		post := Post{
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.False(t, post.CreatedAt.IsZero())
		assert.False(t, post.UpdatedAt.IsZero())
		assert.True(t, post.CreatedAt.Equal(now))
		assert.True(t, post.UpdatedAt.Equal(now))
	})

	t.Run("PostMediaRelation", func(t *testing.T) {
		media1 := Media{ID: 1, URL: "https://example.com/image1.jpg", Type: "image"}
		media2 := Media{ID: 2, URL: "https://example.com/image2.jpg", Type: "image"}

		post := Post{
			ID:      1,
			Title:   "Post with Media",
			Content: "This post has media attachments",
			Media:   []Media{media1, media2},
		}

		assert.Len(t, post.Media, 2)
		assert.Equal(t, "https://example.com/image1.jpg", post.Media[0].URL)
		assert.Equal(t, "https://example.com/image2.jpg", post.Media[1].URL)
		assert.Equal(t, "image", post.Media[0].Type)
		assert.Equal(t, "image", post.Media[1].Type)
	})
}

func TestPostMediaModel(t *testing.T) {
	t.Run("PostMediaCreation", func(t *testing.T) {
		postMedia := PostMedia{
			PostID:  1,
			MediaID: 2,
		}

		assert.Equal(t, uint(1), postMedia.PostID)
		assert.Equal(t, uint(2), postMedia.MediaID)
	})
}
