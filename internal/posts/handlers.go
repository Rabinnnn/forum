package posts

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"forum/internal/auth"
)

type PostHandler struct {
	service   *PostService
	templates *template.Template
}

func NewPostHandler(service *PostService, templates *template.Template) *PostHandler {
	return &PostHandler{
		service:   service,
		templates: templates,
	}
}

func (h *PostHandler) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreatePostHandler called with method: %s", r.Method)

	// Check if user is logged in
	userID, ok := auth.GetUserIDFromSession(r)
	if !ok {
		log.Println("User not logged in, redirecting to login")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// GET request - show create post form
	if r.Method == http.MethodGet {
		data := map[string]interface{}{
			"IsLoggedIn": true,
			"UserID":     userID,
		}

		if err := h.templates.ExecuteTemplate(w, "createPost.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Error rendering page", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Parse the form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		post := &Post{
			UserID:  userID,
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		}

		// Handle image upload if present
		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			uploadDir := filepath.Join("internal", "web", "static", "uploads")
			if err := os.MkdirAll(uploadDir, 0755); err != nil {
				log.Printf("Failed to create uploads directory: %v", err)
				http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
				return
			}

			filename := filepath.Join(uploadDir, handler.Filename)
			f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				log.Printf("Error saving file: %v", err)
				http.Error(w, "Error saving file", http.StatusInternalServerError)
				return
			}
			defer f.Close()

			post.ImagePath = filepath.Join("uploads", handler.Filename)
		}

		if err := h.service.CreatePost(post); err != nil {
			log.Printf("Error creating post: %v", err)
			http.Error(w, "Error creating post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

func (h *PostHandler) GetAllPostsHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

func (h *PostHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	log.Printf("ServeHome called: %s %s", r.Method, r.URL.Path)

	// Get user session
	userID, isLoggedIn := auth.GetUserIDFromSession(r)
	log.Printf("Session state: userID=%s, isLoggedIn=%v", userID, isLoggedIn)

	// Get posts
	posts, err := h.service.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"IsLoggedIn":    isLoggedIn,
		"CurrentUserID": userID,
		"Posts":         posts,
	}

	log.Printf("Rendering home template with data: %+v", data)

	// Execute template
	err = h.templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
