package comments

import (
	"database/sql"
	"encoding/json"
	//"fmt"
	"forum/internal/db"
	"log"
	"net/http"
	//"strconv"
	"strings"

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
	currentUserID, ok := auth.GetUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	comments, totalCount, err := h.service.GetComments(postID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	for i := range comments {
		comments[i].CurrentUserID = currentUserID
	}
	// Wrap comments and total count in a structured response
	response := map[string]interface{}{
		"total_count": totalCount,
		"comments":    comments,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)
}


func HandleEditComment(w http.ResponseWriter, r *http.Request) {
	
	userID, _ := auth.GetUserIDFromSession(r)

	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID := r.FormValue("comment_id")
	// commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	// if err != nil {
	// 	http.Error(w, "Invalid comment ID", http.StatusBadRequest)
	// 	return
	// }

	newContent := strings.TrimSpace(r.FormValue("content"))
	if newContent == "" {
		http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		return
	}

	// Ensure user owns the comment
	var ownerID string
	err := db.Globaldb.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking comment ownership: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Update the comment
	result, err := db.Globaldb.Exec("UPDATE comments SET content = ? WHERE id = ?", newContent, commentID)
	if err != nil {
		log.Printf("Error updating comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
	} else if rowsAffected == 0 {
		log.Printf("No rows were updated for comment ID: %d", commentID)
	}


	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	//http.Redirect(w, r, "/", http.StatusSeeOther) // Redirects to the homepage or correct page

}


func HandleDeleteComment(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromSession(r)
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	commentID := r.FormValue("comment_id")

	// commentID, err := strconv.Atoi(r.FormValue("comment_id"))
	// if err != nil {
	// 	http.Error(w, "Invalid comment ID", http.StatusBadRequest)
	// 	return
	// }

	// Get the post ID before deleting the comment (for updating comment count)
	var postID int
	err := db.Globaldb.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting post ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Ensure user owns the comment
	var ownerID string
	err = db.Globaldb.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking comment ownership: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete the comment
	_, err = db.Globaldb.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		log.Printf("Error deleting comment: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the post's comment count
	if postID > 0 {
		_, err = db.Globaldb.Exec(`
			UPDATE posts 
			SET comments = (
				SELECT COUNT(*) 
				FROM comments 
				WHERE post_id = ?
			) 
			WHERE id = ?`, postID, postID)
		if err != nil {
			log.Printf("Error updating post comment count: %v", err)
		}
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}