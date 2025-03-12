package auth // handle authentication

import (
	"html/template"
	"log"
	"net/http"
)

// User structure
type User struct {
	UserName string
	Email    string
	Password string
}

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

// Serve registration page
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("RegisterHandler called with method:", r.Method)

	if r.Method == http.MethodGet {
		log.Println("Serving registration page")
		tmpl, err := template.ParseFiles("internal/web/templates/register.html")
		if err != nil {
			log.Println("Template error:", err)
			http.Error(w, "Template not found", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Processing registration form")
		r.ParseForm()
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm-password")

		errors := Error{}
		user := User{UserName: username, Email: email}

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
			tmpl, err := template.ParseFiles("internal/web/templates/register.html")
			if err != nil {
				log.Println("Template error:", err)
				http.Error(w, "Template not found", http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, PageData{User: user, Errors: errors})
			return
		}

		log.Println("New user registered:", username, email)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
