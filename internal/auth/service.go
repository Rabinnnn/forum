package auth

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
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

// LoginHandler handles the login route
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler called with method: %s", r.Method)

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		var user User
		err := h.db.QueryRow(
			"SELECT id, username, email, password FROM users WHERE username = ? OR email = ?",
			username, username,
		).Scan(&user.ID, &user.Username, &user.Email, &user.Password)

		if err != nil {
			if err == sql.ErrNoRows {
				data := map[string]interface{}{
					"Error":    "Invalid credentials",
					"Username": username,
				}
				h.templates.ExecuteTemplate(w, "login.html", data)
				return
			}
			log.Printf("Database error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// TODO: Replace with proper password comparison using bcrypt
		if password == user.Password {
			CreateSession(w, user.ID)
			log.Printf("Login successful for user: %s (%s)", user.Username, user.ID)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		data := map[string]interface{}{
			"Error":    "Invalid credentials",
			"Username": username,
		}
		h.templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// RegisterHandler handles the register route
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("RegisterHandler called with method: %s", r.Method)

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		// Validate input
		if password != confirmPassword {
			data := map[string]interface{}{
				"Error":    "Passwords do not match",
				"Username": username,
				"Email":    email,
			}
			h.templates.ExecuteTemplate(w, "register.html", data)
			return
		}

		// Create new user
		userID := uuid.New().String()
		_, err := h.db.Exec(
			"INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)",
			userID, username, email, password, // TODO: Hash password before storing
		)

		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				data := map[string]interface{}{
					"Error":    "Username or email already exists",
					"Username": username,
					"Email":    email,
				}
				h.templates.ExecuteTemplate(w, "register.html", data)
				return
			}
			log.Printf("Database error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Automatically log in the new user
		CreateSession(w, userID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LogoutHandler handles the logout route
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LogoutHandler called with method: %s", r.Method)

	ClearSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL path
	userID := strings.TrimPrefix(r.URL.Path, "/profile/")
	if userID == "" {
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

	// Get current user session
	currentUserID, isLoggedIn := GetUserIDFromSession(r)

	// Prepare template data
	data := map[string]interface{}{
		"IsLoggedIn":    isLoggedIn,
		"CurrentUserID": currentUserID,
		"User":          profileUser,
		"Username":      profileUser.Username,
		"Email":         profileUser.Email,
		"ProfilePic":    profileUser.ProfilePic,
	}

	// Execute template
	if err := h.templates.ExecuteTemplate(w, "profile.html", data); err != nil {
		log.Printf("Error executing profile template: %v", err)
		http.Error(w, "Error rendering profile", http.StatusInternalServerError)
		return
	}
}
