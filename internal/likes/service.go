package likes

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
)

type LikeService struct {
	db *sql.DB
}

func NewLikeService(db *sql.DB) *LikeService {
	return &LikeService{db: db}
}

func (s *LikeService) ToggleLike(userID string, postID string, isLike bool) (*LikeResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Check if like already exists
	var existingLike Like
	err = tx.QueryRow(
		"SELECT id, is_like FROM likes WHERE user_id = ? AND post_id = ?",
		userID, postID,
	).Scan(&existingLike.ID, &existingLike.IsLike)

	if err == sql.ErrNoRows {
		// Create new like with UUID
		newID := uuid.New().String()
		_, err = tx.Exec(
			"INSERT INTO likes (id, user_id, post_id, is_like) VALUES (?, ?, ?, ?)",
			newID, userID, postID, isLike,
		)
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing like: %v", err)
	} else {
		// Update existing like if different
		if existingLike.IsLike != isLike {
			_, err = tx.Exec(
				"UPDATE likes SET is_like = ? WHERE id = ?",
				isLike, existingLike.ID,
			)
		} else {
			// Remove like if same reaction
			_, err = tx.Exec("DELETE FROM likes WHERE id = ?", existingLike.ID)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update like: %v", err)
	}

	// Get updated counts
	var likes, dislikes int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 1",
		postID,
	).Scan(&likes)
	if err != nil {
		return nil, fmt.Errorf("failed to count likes: %v", err)
	}

	err = tx.QueryRow(
		"SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 0",
		postID,
	).Scan(&dislikes)
	if err != nil {
		return nil, fmt.Errorf("failed to count dislikes: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &LikeResponse{
		Likes:    likes,
		Dislikes: dislikes,
	}, nil
}
