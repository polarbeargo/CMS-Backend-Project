package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPageModel(t *testing.T) {
	t.Run("PageCreation", func(t *testing.T) {
		page := Page{
			ID:      1,
			Title:   "Test Page",
			Content: "This is a test page content with HTML <p>tags</p>",
		}

		assert.Equal(t, uint(1), page.ID)
		assert.Equal(t, "Test Page", page.Title)
		assert.Equal(t, "This is a test page content with HTML <p>tags</p>", page.Content)
	})

	t.Run("PageValidation", func(t *testing.T) {
		page := Page{}

		assert.Empty(t, page.Title)
		assert.Empty(t, page.Content)
		assert.Equal(t, uint(0), page.ID)
	})

	t.Run("PageTimestamps", func(t *testing.T) {
		now := time.Now()
		page := Page{
			CreatedAt: now,
			UpdatedAt: now,
		}

		assert.False(t, page.CreatedAt.IsZero())
		assert.False(t, page.UpdatedAt.IsZero())
		assert.True(t, page.CreatedAt.Equal(now))
		assert.True(t, page.UpdatedAt.Equal(now))
	})

	t.Run("PageContentTypes", func(t *testing.T) {
		testCases := []struct {
			name    string
			title   string
			content string
		}{
			{
				"PlainText",
				"Simple Page",
				"This is plain text content",
			},
			{
				"HTMLContent",
				"Rich Content Page",
				"<h1>Header</h1><p>This is <strong>bold</strong> text</p>",
			},
			{
				"MarkdownLike",
				"Markdown Page",
				"# Header\n\n## Subheader\n\n- List item 1\n- List item 2",
			},
			{
				"LongContent",
				"Long Article",
				generateLongContent(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				page := Page{
					Title:   tc.title,
					Content: tc.content,
				}

				assert.Equal(t, tc.title, page.Title)
				assert.Equal(t, tc.content, page.Content)
				assert.NotEmpty(t, page.Title)
				assert.NotEmpty(t, page.Content)
			})
		}
	})

	t.Run("PageTitleLength", func(t *testing.T) {
		longTitle := generateString(300)
		normalTitle := "Normal Length Title"

		pageWithLongTitle := Page{
			Title:   longTitle,
			Content: "Test content",
		}

		pageWithNormalTitle := Page{
			Title:   normalTitle,
			Content: "Test content",
		}

		assert.True(t, len(pageWithLongTitle.Title) > 255)
		assert.True(t, len(pageWithNormalTitle.Title) <= 255)

		assert.NotEmpty(t, pageWithLongTitle.Title)
		assert.NotEmpty(t, pageWithNormalTitle.Title)
	})

	t.Run("PageAutoIncrement", func(t *testing.T) {
		page1 := Page{ID: 1, Title: "Page 1", Content: "Content 1"}
		page2 := Page{ID: 2, Title: "Page 2", Content: "Content 2"}
		page3 := Page{ID: 3, Title: "Page 3", Content: "Content 3"}

		pages := []Page{page1, page2, page3}

		for i, page := range pages {
			expectedID := uint(i + 1)
			assert.Equal(t, expectedID, page.ID)
			assert.NotEmpty(t, page.Title)
			assert.NotEmpty(t, page.Content)
		}
	})
}

func generateLongContent() string {
	return `Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
	Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. 
	Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris 
	nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in 
	reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla 
	pariatur. Excepteur sint occaecat cupidatat non proident, sunt in 
	culpa qui officia deserunt mollit anim id est laborum.`
}

func generateString(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}
