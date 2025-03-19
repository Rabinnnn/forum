package comments

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	ProfilePic sql.NullString `json:"profile_pic"`
}
