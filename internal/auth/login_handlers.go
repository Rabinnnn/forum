package auth

import (
	"database/sql"
	"forum/internal/db"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates = template.Must(template.ParseGlob(filepath.Join("internal", "web", "templates", "*.html")))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler called with method: %s", r.Method)

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		
		log.Printf("Login attempt for username: %s", username)

		// Get database connection
		db, err := db.InitializeDB()
		if err != nil {
			log.Printf("Database initialization error: %v", err)
			data := map[string]interface{}{
				"Error": "Internal server error",
				"Username": username,
			}
			templates.ExecuteTemplate(w, "login.html", data)
			return
		}
		defer db.Close()

		// Query user
		var user User
		err = db.QueryRow(
			"SELECT id, username, email, password FROM users WHERE username = ? OR email = ?",
			username, username,
		).Scan(&user.ID, &user.Username, &user.Email, &user.Password)

		if err != nil {
			log.Printf("Database query error: %v", err)
			if err == sql.ErrNoRows {
				data := map[string]interface{}{
					"Error": "Invalid credentials",
					"Username": username,
				}
				templates.ExecuteTemplate(w, "login.html", data)
				return
			}
			// Handle other errors...
			return
		}

		// Verify password (replace with bcrypt in production)
		if password == user.Password {
			log.Printf("Login successful for user: %s", username)
			
			// Create session
			CreateSession(w, user.ID)
			
			log.Printf("Session created for user: %s", username)
			
			// Redirect to home
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		log.Printf("Invalid password for user: %s", username)
		data := map[string]interface{}{
			"Error": "Invalid credentials",
			"Username": username,
		}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}

	// GET request - show login form
	templates.ExecuteTemplate(w, "login.html", nil)
}
