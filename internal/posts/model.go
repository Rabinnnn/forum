package posts

import (
	"database/sql"
	"time"

	"forum/internal/comments"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Post struct {
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	Title      string             `json:"title"`
	Content    string             `json:"content"`
	Categories []Category         `json:"categories"`
	ImagePath  string             `json:"image_path"`
	CreatedAt  time.Time          `json:"created_at"`
	Username   string             `json:"username"`
	ProfilePic sql.NullString     `json:"profile_pic"`
	Likes      int                `json:"likes"`
	Dislikes   int                `json:"dislikes"`
	Comments   []comments.Comment `json:"comments"`
	PostTime   string             `json:"post_time,omitempty"`
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
