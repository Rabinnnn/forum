package auth

import (
	"database/sql"
	"testing"
	"time"
)

func TestUserStructInitialization(t *testing.T) {
	user := User{
		ID:         "u123",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpassword",
		ProfilePic: sql.NullString{String: "path/to/pic.jpg", Valid: true},
	}

	if user.Username != "testuser" {
		t.Errorf("expected Username 'testuser', got '%s'", user.Username)
	}

	if !user.ProfilePic.Valid {
		t.Error("expected ProfilePic to be valid")
	}
}

func TestPostStructInitialization(t *testing.T) {
	now := time.Now()
	post := Post{
		ID:        "p123",
		Title:     "Hello World",
		Content:   "This is a test post.",
		ImagePath: "/images/test.jpg",
		CreatedAt: now,
		Likes:     10,
		Dislikes:  2,
		Comments:  3,
	}

	if post.Likes != 10 {
		t.Errorf("expected 10 likes, got %d", post.Likes)
	}

	if post.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

// helper to check substring in string
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && s[len(substr)-1] == substr[len(substr)-1])
}
