package auth

import (
	"testing"
	"time"
)

func TestPost(t *testing.T) {
	tests := []struct {
		name     string
		post     Post
		expected Post
	}{
		{
			name: "Valid Post",
			post: Post{
				ID:        "1",
				Title:     "Test Title",
				Content:   "Test Content",
				ImagePath: "/images/test.jpg",
				CreatedAt: time.Now(),
				Likes:     10,
				Dislikes:  2,
				Comments:  5,
			},
			expected: Post{
				ID:        "1",
				Title:     "Test Title",
				Content:   "Test Content",
				ImagePath: "/images/test.jpg",
				Likes:     10,
				Dislikes:  2,
				Comments:  5,
			},
		},
		{
			name: "Empty Post",
			post: Post{},
			expected: Post{
				ID:        "",
				Title:     "",
				Content:   "",
				ImagePath: "",
				Likes:     0,
				Dislikes:  0,
				Comments:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.post.ID != tt.expected.ID {
				t.Errorf("expected ID %s, got %s", tt.expected.ID, tt.post.ID)
			}
			if tt.post.Title != tt.expected.Title {
				t.Errorf("expected Title %s, got %s", tt.expected.Title, tt.post.Title)
			}
			if tt.post.Content != tt.expected.Content {
				t.Errorf("expected Content %s, got %s", tt.expected.Content, tt.post.Content)
			}
			if tt.post.ImagePath != tt.expected.ImagePath {
				t.Errorf("expected ImagePath %s, got %s", tt.expected.ImagePath, tt.post.ImagePath)
			}
			if tt.post.Likes != tt.expected.Likes {
				t.Errorf("expected Likes %d, got %d", tt.expected.Likes, tt.post.Likes)
			}
			if tt.post.Dislikes != tt.expected.Dislikes {
				t.Errorf("expected Dislikes %d, got %d", tt.expected.Dislikes, tt.post.Dislikes)
			}
			if tt.post.Comments != tt.expected.Comments {
				t.Errorf("expected Comments %d, got %d", tt.expected.Comments, tt.post.Comments)
			}
		})
	}
}
