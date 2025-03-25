package comments

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID         string         `json:"id"`
	PostID     string         `json:"post_id"`
	UserID     string         `json:"user_id"`
	CurrentUserID     string         `json:"current_user_id"`
	Content    string         `json:"content"`
	Likes      int			  `json:"likes"`
	Dislikes   int            `json:"dislikes"`
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
func (c Comment) GetLikes() int      { return c.Likes }
func (c Comment) GetDislikes() int      { return c.Dislikes }
