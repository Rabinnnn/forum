package auth

import (
	"html/template"
	"net/http"
	"path/filepath"
)

var templates = template.Must(template.ParseGlob(filepath.Join("internal", "web", "templates", "*.html")))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Handle login logic (e.g., authenticate user)
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "admin" && password == "password" {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		data := map[string]string{"GeneralError": "Invalid credentials", "Username": username}
		templates.ExecuteTemplate(w, "login.html", data)
		return
	}
	templates.ExecuteTemplate(w, "login.html", nil)
}
