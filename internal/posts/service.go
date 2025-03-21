package posts

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"forum/internal/comments"

	//"github.com/google/uuid"
	"math/rand"
)

type PostService struct {
	db *sql.DB
}

func NewPostService(db *sql.DB) *PostService {
	return &PostService{db: db}
}

func GeneratePostID() int {
    rand.Seed(time.Now().UnixNano()) // Ensure different values on each run
    return rand.Intn(1000000) + 1 // Generate a random ID between 1 and 1,000,000
}
func (s *PostService) CreatePost(post *Post) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Generate new UUID for post
	//post.ID = uuid.New().String()
	postInt := GeneratePostID() // If using integer-based IDs
	postStr := strconv.Itoa(postInt)
	post.ID = postStr

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

func (s *PostService) GetAllPosts() ([]Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.image_path, p.created_at,
			COALESCE(u.username, 'Unknown') as username,
			u.profile_pic,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 1), 0) as likes,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 0), 0) as dislikes
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
	commentService := comments.NewCommentService(s.db)

	for rows.Next() {
		var post Post
		var profilePic sql.NullString
		err := rows.Scan(
			&post.ID, &post.UserID, &post.Title, &post.Content, &post.ImagePath,
			&post.CreatedAt, &post.Username, &profilePic,
			&post.Likes, &post.Dislikes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %v", err)
		}
		post.ProfilePic = profilePic
		post.PostTime = formatPostTime(post.CreatedAt)

		// Fetch full comment details
		commentsList, err := commentService.GetComments(post.ID)
		if err != nil {
			return nil, err
		}
		post.Comments = make([]CommentData, len(commentsList))
		for i, comment := range commentsList {
			post.Comments[i] = comment // This works because Comment implements CommentData
		}

		// Fetch categories
		categories, err := s.getPostCategories(post.ID)
		if err != nil {
			return nil, err
		}
		post.Categories = categories // Store categories

		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostService) GetUserPosts(userID string) ([]Post, error) {
	query := `
		SELECT 
			p.id, p.title, p.content, p.image_path, p.created_at,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 1), 0) as likes,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND like = 0), 0) as dislikes,
			COALESCE((SELECT COUNT(*) FROM comments WHERE post_id = p.id), 0) as comments
		FROM posts p
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user posts: %v", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(
			&post.ID, &post.Title, &post.Content, &post.ImagePath,
			&post.CreatedAt, &post.Likes, &post.Dislikes, &post.Comments,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %v", err)
		}
		post.PostTime = formatPostTime(post.CreatedAt)
		posts = append(posts, post)
	}

	return posts, nil
}

func (s *PostService) getPostCategories(postID string) ([]Category, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.name 
		FROM categories c
		JOIN post_categories pc ON c.id = pc.category_id
		WHERE pc.post_id = ?
		ORDER BY c.name
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories for post %s: %v", postID, err)
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

func formatPostTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	default:
		return t.Format("Jan 2")
	}
}
