package auth

import (
	"database/sql"
	"testing"
	"time"
)

func TestUserStruct(t *testing.T) {
	user := User{
		ID:         "1",
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpassword",
		ProfilePic: sql.NullString{String: "pic.jpg", Valid: true},
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username to be 'testuser', got '%s'", user.Username)
	}

	if !user.ProfilePic.Valid {
		t.Error("Expected ProfilePic to be valid")
	}
}

func TestPostStruct(t *testing.T) {
	now := time.Now()
	post := Post{
		ID:        "post1",
		Title:     "Test Title",
		Content:   "Test Content",
		ImagePath: "/images/test.jpg",
		CreatedAt: now,
		Likes:     10,
		Dislikes:  1,
		Comments:  5,
	}

	if post.Likes != 10 {
		t.Errorf("Expected 10 likes, got %d", post.Likes)
	}

	if post.CreatedAt != now {
		t.Error("CreatedAt timestamp does not match expected value")
	}
}
