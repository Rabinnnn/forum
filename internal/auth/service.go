package auth

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

// GetUserByID retrieves a user from the database by their ID
func GetUserByID(db *sql.DB, userID string) (*User, error) {
	log.Printf("Fetching user with ID: %s", userID)

	var user User
	var profilePic sql.NullString
	err := db.QueryRow(
		"SELECT id, username, email, COALESCE(profile_pic, '') as profile_pic FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &profilePic)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No user found with ID: %s", userID)
			return nil, nil
		}
		log.Printf("Database error while fetching user %s: %v", userID, err)
		return nil, err
	}

	user.ProfilePic = profilePic
	log.Printf("Successfully fetched user: %s (%s)", user.Username, user.ID)
	return &user, nil
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(db *sql.DB, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		db:        db,
		templates: templates,
	}
}

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db        *sql.DB
	templates *template.Template
}

// LogoutHandler handles the logout route
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LogoutHandler called with method: %s", r.Method)

	ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ProfileHandler called with path: %s", r.URL.Path)

	// Get user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/profile/")
	log.Printf("Extracted userID: %q", userID)

	if userID == "" || userID == "/" {
		log.Printf("Invalid userID, returning 404")
		http.NotFound(w, r)
		return
	}

	// Get profile user
	profileUser, err := GetUserByID(h.db, userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	if profileUser == nil {
		log.Printf("No user found with ID: %s", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	log.Printf("Found user: %+v", profileUser)

	// Get current user session
	currentUserID, isLoggedIn := GetUserIDFromSession(r)
	log.Printf("Current user session: userID=%s, isLoggedIn=%v", currentUserID, isLoggedIn)

	// Get user's posts
	posts, err := h.getUserPosts(userID)
	if err != nil {
		log.Printf("Error getting user posts: %v", err)
		posts = []Post{} // Continue without posts rather than failing
	}
	log.Printf("Found %d posts for user", len(posts))

	// Prepare template data
	data := map[string]interface{}{
		"IsLoggedIn":    isLoggedIn,
		"CurrentUserID": currentUserID,
		"User":          profileUser,
		"Username":      profileUser.Username,
		"Email":         profileUser.Email,
		"ProfilePic":    profileUser.ProfilePic,
		"Posts":         posts,
	}

	log.Printf("Rendering template with data: %+v", data)

	// Execute template
	if err := h.templates.ExecuteTemplate(w, "profile.html", data); err != nil {
		log.Printf("Error executing profile template: %v", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
		return
	}
}

// Add this helper method to get user's posts
func (h *AuthHandler) getUserPosts(userID string) ([]Post, error) {
	query := `
		SELECT 
			p.id, p.title, p.content, p.image_path, p.created_at,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 1), 0) as likes,
			COALESCE((SELECT COUNT(*) FROM likes WHERE post_id = p.id AND is_like = 0), 0) as dislikes,
			COALESCE((SELECT COUNT(*) FROM comments WHERE post_id = p.id), 0) as comments
		FROM posts p
		WHERE p.user_id = ?
		ORDER BY p.created_at DESC
	`

	rows, err := h.db.Query(query, userID)
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
		posts = append(posts, post)
	}

	return posts, nil
}
