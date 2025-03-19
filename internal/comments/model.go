package comments

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID         string         `json:"id"`
	PostID     string         `json:"post_id"`
	UserID     string         `json:"user_id"`
	Content    string         `json:"content"`
	CreatedAt  time.Time      `json:"created_at"`
	Username   string         `json:"username"`
	ProfilePic sql.NullString `json:"profile_pic"`
}

// Implement interface methods
func (c Comment) GetID() string { return c.ID }
func (c Comment) GetContent() string { return c.Content }
func (c Comment) GetUserID() string { return c.UserID }
func (c Comment) GetCreatedAt() time.Time { return c.CreatedAt }
func (c Comment) GetUsername() string { return c.Username }
func (c Comment) GetProfilePic() sql.NullString { return c.ProfilePic }
