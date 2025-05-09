package posts

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"forum/internal/xerrors"

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
		// Get categories for the form
		categories, err := h.service.GetCategories()
		if err != nil {
			log.Printf("Error fetching categories: %v", err)
			// http.Error(w, "Error fetching categories", http.StatusInternalServerError)
			xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )

			return
		}

		data := map[string]interface{}{
			"IsLoggedIn":  true,
			"UserID":      userID,
			"Categories":  categories,
		}

		if err := h.templates.ExecuteTemplate(w, "createPost.html", data); err != nil {
			log.Printf("Error executing template: %v", err)
			// http.Error(w, "Error rendering page", http.StatusInternalServerError)
			xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )

		}
		return
	}

	if r.Method == http.MethodPost {
		// Parse the form
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Printf("Error parsing form: %v", err)
			// http.Error(w, "Error parsing form", http.StatusBadRequest)
			xerrors.RenderErrorPage(w, http.StatusBadRequest, xerrors.ErrBadRequest )

			return
		}

		post := &Post{
			UserID:  userID,
			Title:   r.FormValue("title"),
			Content: r.FormValue("content"),
		}

		// Handle categories
		if categories := r.Form["categories[]"]; len(categories) > 0 {
			// for _, catID := range categories {
			// 	post.Categories = append(post.Categories, Category{ID: parseInt(catID)})
			// }
			for _, catID := range categories {
				id, err := strconv.Atoi(catID)
				if err != nil {
					log.Printf("invalid category ID: %v", catID)
					continue // skip invalid IDs
				}
				post.Categories = append(post.Categories, Category{ID: id})
			}
			fmt.Printf("Parsed categories for post: %+v\n", post.Categories)

			
		}

		// Handle image upload if present
		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Change the upload directory to be relative to the static directory
			uploadDir := filepath.Join("internal", "web", "static", "uploads")
			if err := os.MkdirAll(uploadDir, 0o755); err != nil {
				log.Printf("Failed to create uploads directory: %v", err)
				// http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
				xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )
				return
			}

			// Generate a unique filename to prevent overwrites
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), handler.Filename)
			fullPath := filepath.Join(uploadDir, filename)

			f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE, 0o666)
			if err != nil {
				log.Printf("Error saving file: %v", err)
				// http.Error(w, "Error saving file", http.StatusInternalServerError)
				xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )

				return
			}
			defer f.Close()

			io.Copy(f, file)

			// Store the path relative to the static directory
			post.ImagePath = "/static/uploads/" + filename
		}

		if err := h.service.CreatePost(post); err != nil {
			log.Printf("Error creating post: %v", err)
			// http.Error(w, "Error creating post", http.StatusInternalServerError)
			xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )
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
		// http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )
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
		// http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )
		return
	}

	// For each post, we need to set IsLoggedIn
	for i := range posts {
		posts[i].IsLoggedIn = isLoggedIn
	}

	// Prepare template data
	data := map[string]interface{}{
		"IsLoggedIn": isLoggedIn,
		"UserID":     userID,
		"Posts":      posts,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		// http.Error(w, "Error rendering page", http.StatusInternalServerError)
		xerrors.RenderErrorPage(w, http.StatusInternalServerError, xerrors.ErrInternalServer )

		return
	}
}

func (h *PostHandler) TestPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "test_posts.html", nil)
		return
	}
}

// Helper function to parse string to int
func parseInt(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}
