package posts

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type PostService struct {
	db *sql.DB
}

func NewPostService(db *sql.DB) *PostService {
	return &PostService{db: db}
}

func (s *PostService) CreatePost(post *Post) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Generate new UUID for post
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()

	// Insert post
	_, err = tx.Exec(`
		INSERT INTO posts (id, user_id, title, content, image_path, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, post.ID, post.UserID, post.Title, post.Content, post.ImagePath, post.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert post: %v", err)
	}

	// Insert categories
	for _, category := range post.Categories {
		_, err = tx.Exec(`
			INSERT INTO post_categories (post_id, category_id)
			VALUES (?, ?)
		`, post.ID, category.ID)
		if err != nil {
			return fmt.Errorf("failed to insert category %d: %v", category.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// New method to get all categories
func (s *PostService) GetCategories() ([]Category, error) {
	rows, err := s.db.Query("SELECT id, name FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %v", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		if err := rows.Scan(&cat.ID, &cat.Name); err != nil {
			return nil, fmt.Errorf("failed to scan category: %v", err)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

// Modified GetAllPosts to include category information
func (s *PostService) GetAllPosts() ([]Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.image_path, p.created_at,
			COALESCE(u.username, 'Unknown') as username,
			COALESCE(u.profile_pic, '') as profile_pic,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1), 0) as likes,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0), 0) as dislikes,
			COALESCE((SELECT COUNT(*) FROM comments WHERE post_id = p.id), 0) as comments
		FROM posts p
		LEFT JOIN users u ON p.user_id = u.id
		ORDER BY p.created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var profilePic sql.NullString
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImagePath,
			&post.CreatedAt, &post.Username, &profilePic,
			&post.Likes, &post.Dislikes, &post.Comments,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %v", err)
		}
		post.ProfilePic = profilePic

		// Get categories for this post
		catRows, err := s.db.Query(`
			SELECT c.id, c.name 
			FROM categories c
			JOIN post_categories pc ON c.id = pc.category_id
			WHERE pc.post_id = ?
			ORDER BY c.name
		`, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query categories for post %s: %v", post.ID, err)
		}
		defer catRows.Close()

		for catRows.Next() {
			var cat Category
			if err := catRows.Scan(&cat.ID, &cat.Name); err != nil {
				return nil, fmt.Errorf("failed to scan category: %v", err)
			}
			post.Categories = append(post.Categories, cat)
		}

		posts = append(posts, post)
	}

	return posts, nil
}
