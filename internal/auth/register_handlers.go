package auth

import (
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Error structure for validation messages
type Error struct {
	GeneralError  string
	UsernameError string
	EmailError    string
	PasswordError string
}

// PageData to hold form data and errors
type PageData struct {
	User   User
	Errors Error
}

func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

		errors := Error{}
		user := User{
			Username: username,
			Email:    email,
		}

		// Basic validation
		if username == "" {
			errors.UsernameError = "Username is required."
		}
		if email == "" {
			errors.EmailError = "Email is required."
		}
		if password == "" {
			errors.PasswordError = "Password is required."
		} else if password != confirmPassword {
			errors.PasswordError = "Passwords do not match."
		}

		if errors.UsernameError != "" || errors.EmailError != "" || errors.PasswordError != "" {
			log.Println("Validation errors:", errors)
			h.templates.ExecuteTemplate(w, "register.html", PageData{User: user, Errors: errors})
			return
		}

		// Hash the password before storing
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			errors.GeneralError = "Internal server error"
			h.templates.ExecuteTemplate(w, "register.html", PageData{User: user, Errors: errors})
			return
		}

		// Insert the new user with hashed password
		userID := uuid.New().String()
		_, err = h.db.Exec(
			"INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)",
			userID,
			username,
			email,
			string(hashedPassword), // Store the hashed password
		)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				errors.GeneralError = "Username or email already exists"
				h.templates.ExecuteTemplate(w, "register.html", PageData{User: user, Errors: errors})
				return
			}
			log.Printf("Database error: %v", err)
			errors.GeneralError = "Internal server error"
			h.templates.ExecuteTemplate(w, "register.html", PageData{User: user, Errors: errors})
			return
		}

		// Automatically log in the new user
		CreateSession(w, userID)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}
