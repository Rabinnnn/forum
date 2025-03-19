package comments

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"forum/internal/auth"
)

type CommentHandler struct {
	service *CommentService
}

func NewCommentHandler(db *sql.DB) *CommentHandler {
	return &CommentHandler{
		service: NewCommentService(db),
	}
}

func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	userID, ok := auth.GetUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate input
	if input.Content == "" || input.PostID == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Create the comment
	if err := h.service.AddComment(input.PostID, userID, input.Content); err != nil {
		http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Comment created successfully",
	})
}

func (h *CommentHandler) GetCommentsForPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	comments, err := h.service.GetComments(postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comments)
}
