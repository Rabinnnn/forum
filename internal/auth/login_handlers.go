package auth

import (
	"database/sql"
	"forum/internal/db"
	"html/template"
	"net/http"
	"path/filepath"
)

var templates = template.Must(template.ParseGlob(filepath.Join("internal", "web", "templates", "*.html")))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		// Get database connection
		db, err := db.InitializeDB()
		if err != nil {
			data := map[string]string{"GeneralError": "Internal server error", "Username": username}
			templates.ExecuteTemplate(w, "login.html", data)
			return
		}
		defer db.Close()

		// Query the database
		var storedPassword string
		err = db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&storedPassword)

		if err == sql.ErrNoRows {
			data := map[string]string{"GeneralError": "Invalid credentials", "Username": username}
			templates.ExecuteTemplate(w, "login.html", data)
			return
		} else if err != nil {
			data := map[string]string{"GeneralError": "Internal server error", "Username": username}
			templates.ExecuteTemplate(w, "login.html", data)
			return
		}

		// Compare passwords using bcrypt
		if password == storedPassword { // Replace with bcrypt.CompareHashAndPassword() in production
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		data := map[string]string{"GeneralError": "Invalid credentials", "Username": username}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}
