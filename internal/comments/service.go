package comments

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type CommentService struct {
	db *sql.DB
}

func NewCommentService(db *sql.DB) *CommentService {
	return &CommentService{db: db}
}

// AddComment adds a new comment to a post
func (s *CommentService) AddComment(postID, userID, content string) error {
	commentID := uuid.New().String()
	createdAt := time.Now()

	_, err := s.db.Exec(`
		INSERT INTO comments (id, post_id, user_id, content, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, commentID, postID, userID, content, createdAt)
	if err != nil {
		return fmt.Errorf("failed to insert comment: %v", err)
	}

	return nil
}

// GetComments retrieves all comments for a given post
func (s *CommentService) GetComments(postID string) ([]Comment, int, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.post_id, c.user_id, c.content, c.created_at, c.likes, c.dislikes,
		u.username, u.profile_pic,
		COUNT(*) OVER() AS total_count  -- Get total count
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query comments: %v", err)
	}
	defer rows.Close()

	var comments []Comment
	var totalCount int
	for rows.Next() {
		var comment Comment
		var profilePic sql.NullString
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.Likes, &comment.Dislikes, &comment.Username, &profilePic, &totalCount)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan comment: %v", err)
		}
		comment.ProfilePic = profilePic
		comments = append(comments, comment)
	}
	// Debugging: Print retrieved comments
	fmt.Printf("Comments for post %s: %+v\n", postID, comments)
	return comments, totalCount, nil
}
