package comments

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"forum/internal/db"
	"forum/internal/xerrors"
	"log"
	"net/http"

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
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		xerrors.RenderErrorPage(w, http.StatusUnauthorized, xerrors.ErrUnauthorized)
		return
	}

	var input struct {
		PostID  string `json:"post_id"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		// http.Error(w, "Invalid request payload", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
		return
	}

	// Validate input
	if input.Content == "" || input.PostID == "" {
		// http.Error(w, "Missing required fields", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
		return
	}

	// Create the comment
	if err := h.service.AddComment(input.PostID, userID, input.Content); err != nil {
		// http.Error(w, "Failed to create comment", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
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
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		xerrors.RenderErrorPage(w, http.StatusUnauthorized, xerrors.ErrUnauthorized)
		return
	}

	postID := r.URL.Query().Get("post_id")
	if postID == "" {
		// http.Error(w, "Missing post ID", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
		return
	}

	comments, totalCount, err := h.service.GetComments(postID)
	if err != nil {
		// http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
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
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		xerrors.RenderErrorPage(w, http.StatusUnauthorized, xerrors.ErrUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		// http.Error(w, "Bad request", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
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
		// http.Error(w, "Comment cannot be empty", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
		return
	}

	// Ensure user owns the comment
	var ownerID string
	err := db.Globaldb.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		log.Printf("CommentID: %v", commentID)
		// http.Error(w, "Comment not foundddd", http.StatusNotFound)
		xerrors.RenderErrorPage(w, http.StatusNotFound, xerrors.ErrNotFound)

		return
	} else if err != nil {
		log.Printf("Error checking comment ownership: %v", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	if ownerID != userID {
		// http.Error(w, "Forbidden", http.StatusForbidden)
		xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrForbidden)
		return
	}

	// Update the comment
	result, err := db.Globaldb.Exec("UPDATE comments SET content = ? WHERE id = ?", newContent, commentID)
	if err != nil {
		log.Printf("Error updating comment: %v", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
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
		// http.Error(w, "Unauthorized", http.StatusUnauthorized)
		xerrors.RenderErrorPage(w, http.StatusUnauthorized, xerrors.ErrUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		// http.Error(w, "Bad request", http.StatusBadRequest)
		xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest)
		return
	}

	commentID := r.FormValue("comment_id")

	// Get the post ID before deleting the comment (for updating comment count)
	var postID int
	err := db.Globaldb.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error getting post ID: %v", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	// Ensure user owns the comment
	var ownerID string
	err = db.Globaldb.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&ownerID)
	if err == sql.ErrNoRows {
		log.Printf("CommentID: %v", commentID)

		// http.Error(w, "Comment notttt found", http.StatusNotFound)
		xerrors.RenderErrorPage(w, http.StatusNotFound, xerrors.ErrNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking comment ownership: %v", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
		return
	}

	if ownerID != userID {
		// http.Error(w, "Forbidden", http.StatusForbidden)
		xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrBadRequest)
		return
	}

	// Delete the comment
	_, err = db.Globaldb.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		log.Printf("Error deleting comment: %v", err)
		// http.Error(w, "Internal server error", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
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



func HandleCommentReactions(w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.GetUserIDFromSession(r)

	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
		return
	}

	var req struct {
		CommentID string `json:"comment_id"`
		Like      int `json:"like"` // 1 for like, 0 for dislike
	}



	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Println("Error:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Like != 0 && req.Like != 1 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid reaction type"})
		return
	}

	// Check if the user already reacted to this comment by querying comment_reaction
	var existingIsLike int
	err := db.Globaldb.QueryRow("SELECT comment_like FROM likes WHERE user_id = ? AND comment_id = ?", userID, req.CommentID).Scan(&existingIsLike)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("Error:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (select)"})
		return
	}

	if err == sql.ErrNoRows {
		// No reaction exists—insert a new reaction
		_, err = db.Globaldb.Exec("INSERT INTO likes (user_id, comment_id, comment_like) VALUES (?, ?, ?)", userID, req.CommentID, req.Like)
		if err != nil {
			fmt.Println("Error:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error (insert)"})
			return
		}
		if req.Like == 1 {
			_, err = db.Globaldb.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", req.CommentID)
		} else {
			_, err = db.Globaldb.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", req.CommentID)
		}
	} else {
		if existingIsLike == req.Like {
			// Same reaction exists; remove it
			_, err = db.Globaldb.Exec("DELETE FROM likes WHERE user_id = ? AND comment_id = ?", userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (delete)"})
				return
			}
			if req.Like == 1{
				_, err = db.Globaldb.Exec("UPDATE comments SET likes = likes - 1 WHERE id = ?", req.CommentID)
				if err != nil {
					http.Error(w, "Database error4.1", http.StatusInternalServerError)
					return
				}
			}else{
				_, err = db.Globaldb.Exec("UPDATE comments SET dislikes = dislikes - 1 WHERE id = ?", req.CommentID)
				if err != nil {
					http.Error(w, "Database error4.2", http.StatusInternalServerError)
					return
				}
			}
		} else {
			// Reaction exists but is different; update it
			_, err = db.Globaldb.Exec("UPDATE likes SET comment_like = ? WHERE user_id = ? AND comment_id = ?", req.Like, userID, req.CommentID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Database error (update)"})
				return
			}

			
			if req.Like == 1{
				_, err = db.Globaldb.Exec("UPDATE comments SET likes = likes + 1 WHERE id = ?", req.CommentID)
				if err != nil {
					// http.Error(w, "Database error4.1", http.StatusInternalServerError)
					xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
					return
				}
				_, err = db.Globaldb.Exec("UPDATE comments SET dislikes = dislikes - 1 WHERE id = ?", req.CommentID)

			}else{
				_, err = db.Globaldb.Exec("UPDATE comments SET dislikes = dislikes + 1 WHERE id = ?", req.CommentID)
				if err != nil {
					// http.Error(w, "Database error4.1", http.StatusInternalServerError)
					xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer)
					return
				}
				_, err = db.Globaldb.Exec("UPDATE comments SET likes = likes - 1 WHERE id = ?", req.CommentID)

			}
		}
	}

	// Get updated likes and dislikes counts
	var likes, dislikes int
	err = db.Globaldb.QueryRow("SELECT likes, dislikes FROM comments WHERE id = ?", req.CommentID).Scan(&likes, &dislikes)
	fmt.Println("req.CommentID: %v", req.CommentID)

	if err != nil {
		fmt.Println("Error:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Database error (get counts)"})
		return
	}

	// Return success response with updated counts
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"likes":    likes,
		"dislikes": dislikes,
	})
}