package likes

import (
	"database/sql"
	"encoding/json"
	//"fmt"
	"forum/internal/auth"
	"forum/internal/db"
	"log"
	"net/http"
)

func HandleReactions(w http.ResponseWriter, r *http.Request) {
	if db.Globaldb == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Decode the request body **before** using req.PostID
	var req struct {
		PostID int `json:"post_id"` // Change to string to match SQLite's TEXT type
		Like   int    `json:"like"`    // 1 for like, 0 for dislike
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate PostID
	if req.PostID < 0 {
		http.Error(w, "Missing post ID", http.StatusBadRequest)
		return
	}

	userID, isLoggedIn := auth.GetUserIDFromSession(r)
	if userID == "" || !isLoggedIn{
		http.Error(w, "Unauthorized: Missing userID", http.StatusUnauthorized)
		return
	}

	// Get logged-in user ID from request context
	// log.Println("Request Context:", r.Context())
	// userIDValue := r.Context().Value("userID")
	// userID, ok := userIDValue.(string)
	// if !ok || userID == "" {
	// 	http.Error(w, "Unauthorized: Missing userID", http.StatusUnauthorized)
	// 	return
	// }

	// Verify if the post exists and get the post owner's user_id
	var postOwnerID string
	err := db.Globaldb.QueryRow("SELECT user_id FROM posts WHERE id = ?", req.PostID).Scan(&postOwnerID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch user ID", http.StatusInternalServerError)
		log.Println("Database error1:", err)
		return
	}

	// Ensure like value is valid (0 or 1)
	if req.Like != 0 && req.Like != 1 {
		http.Error(w, "Invalid reaction type", http.StatusBadRequest)
		return
	}

	// Check if the user already has a reaction
	var existingLike int
	err = db.Globaldb.QueryRow("SELECT like FROM likes WHERE user_id = ? AND post_id = ?", userID, req.PostID).Scan(&existingLike)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error2", http.StatusInternalServerError)
		return
	}

	if err == sql.ErrNoRows {
		// Insert new reaction
		_, err = db.Globaldb.Exec("INSERT INTO likes (user_id, post_id, like) VALUES (?, ?, ?)", userID, req.PostID, req.Like)
		// if err != nil {
		// 	http.Error(w, fmt.Sprintf("Database error3: %v", err), http.StatusInternalServerError)
		// 	return
		// }
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Database error3"})
			return
		}

		if req.Like == 1 {
			_, err = db.Globaldb.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", req.PostID)
		} else {
			_, err = db.Globaldb.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", req.PostID)
		}
	} else {
		if existingLike == req.Like {
			// User is unliking or undisliking
			_, err = db.Globaldb.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ?", userID, req.PostID)
			if err != nil {
				http.Error(w, "Database error4", http.StatusInternalServerError)
				return
			}

		
			if req.Like == 1{
				_, err = db.Globaldb.Exec("UPDATE posts SET likes = likes - 1 WHERE id = ?", req.PostID)
				if err != nil {
					http.Error(w, "Database error4.1", http.StatusInternalServerError)
					return
				}
			}else{
				_, err = db.Globaldb.Exec("UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?", req.PostID)
				if err != nil {
					http.Error(w, "Database error4.2", http.StatusInternalServerError)
					return
				}
			}
			

		} else {
			// Update existing reaction
			_, err = db.Globaldb.Exec("UPDATE likes SET like = ? WHERE user_id = ? AND post_id = ?", req.Like, userID, req.PostID)
			if err != nil {
				http.Error(w, "Database error5", http.StatusInternalServerError)
				return
			}

			if req.Like == 1{
				_, err = db.Globaldb.Exec("UPDATE posts SET likes = likes + 1 WHERE id = ?", req.PostID)
				if err != nil {
					http.Error(w, "Database error4.1", http.StatusInternalServerError)
					return
				}
				_, err = db.Globaldb.Exec("UPDATE posts SET dislikes = dislikes - 1 WHERE id = ?", req.PostID)

			}else{
				_, err = db.Globaldb.Exec("UPDATE posts SET dislikes = dislikes + 1 WHERE id = ?", req.PostID)
				if err != nil {
					http.Error(w, "Database error4.1", http.StatusInternalServerError)
					return
				}
				_, err = db.Globaldb.Exec("UPDATE posts SET likes = likes - 1 WHERE id = ?", req.PostID)

			}
			
			
		}
	}

	// Fetch updated like and dislike counts
	var likes, dislikes int
	err = db.Globaldb.QueryRow("SELECT likes, dislikes FROM posts WHERE id = ?", req.PostID).Scan(&likes, &dislikes)
	if err != nil {
		http.Error(w, "Database error6", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{
		"likes":    likes,
		"dislikes": dislikes,
	})
}
