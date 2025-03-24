package posts

import (
	"database/sql"
	"time"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CommentData interface {
	GetID() string
	GetContent() string
	GetUserID() string
	GetCreatedAt() time.Time
	// GetUsername() string
	// GetProfilePic() sql.NullString
}

type Post struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	Title      string         `json:"title"`
	Content    string         `json:"content"`
	Categories []Category     `json:"categories"`
	ImagePath  string         `json:"image_path"`
	CreatedAt  time.Time      `json:"created_at"`
	Username   string         `json:"username"`
	ProfilePic sql.NullString `json:"profile_pic"`
	Likes      int            `json:"likes"`
	Dislikes   int            `json:"dislikes"`
	Comments   []CommentData  `json:"comments"`
	CommentCount int		  `json:"comment_count"`
	PostTime   string         `json:"post_time,omitempty"`
	IsLoggedIn bool
}

type PostServiceInterface interface {
	CreatePost(post *Post) error
	GetAllPosts() ([]Post, error)
	GetPostByID(id string) (*Post, error)
	UpdatePost(post *Post) error
	DeletePost(id string) error
	GetCategories() ([]Category, error)
	GetUserPosts(userID string) ([]Post, error)
}



// Ensure Comment implements CommentData
var _ CommentData = (*Comment)(nil)

// Comment struct implementing CommentData
type Comment struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	Content    string         `json:"content"`
	CreatedAt  time.Time      `json:"created_at"`
//	Username   string         `json:"username"`
//	ProfilePic sql.NullString `json:"profile_pic"`
}

// Implement CommentData interface methods
func (c Comment) GetID() string          { return c.ID }
func (c Comment) GetContent() string     { return c.Content }
func (c Comment) GetUserID() string      { return c.UserID }
func (c Comment) GetCreatedAt() time.Time { return c.CreatedAt }
// func (c Comment) GetUsername() string    { return c.Username }
// func (c Comment) GetProfilePic() sql.NullString { return c.ProfilePic }