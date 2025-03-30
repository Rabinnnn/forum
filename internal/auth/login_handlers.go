package auth

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"forum/internal/xerrors"


	"golang.org/x/crypto/bcrypt"
)

var templates = template.Must(template.ParseGlob(filepath.Join("internal", "web", "templates", "*.html")))

func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoginHandler called with method: %s, URL: %s", r.Method, r.URL.Path)

	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		log.Printf("Login attempt for username: %s", username)

		var user User
		var hashedPassword string
		err := h.db.QueryRow(
			"SELECT id, username, email, password FROM users WHERE username = ? OR email = ?",
			username, username,
		).Scan(&user.ID, &user.Username, &user.Email, &hashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Printf("No user found with username/email: %s", username)
				data := map[string]interface{}{
					"Error":    "Invalid credentials",
					"Username": username,
				}
				h.templates.ExecuteTemplate(w, "login.html", data)
				return
			}
			log.Printf("Database error: %v", err)
			//http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )
			return
		}

		// Compare the provided password with the stored hash
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		if err != nil {
			log.Printf("Invalid password for user: %s", username)
			data := map[string]interface{}{
				"Error":    "Invalid credentials",
				"Username": username,
			}
			h.templates.ExecuteTemplate(w, "login.html", data)
			return
		}

		// Password is correct, create session
		CreateSession(w, user.ID)
		log.Printf("Login successful for user: %s (ID: %s)", user.Username, user.ID)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	xerrors.RenderErrorPage(w, http.StatusMethodNotAllowed, xerrors.ErrMethodNotAllowed )

}
