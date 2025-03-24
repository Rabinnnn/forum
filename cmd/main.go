package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"forum/internal/auth"
	"forum/internal/comments"
	"forum/internal/db"
	"forum/internal/likes"
	"forum/internal/posts"
)

// Global variables
var (
	templates *template.Template
	database  *sql.DB
)

// Define data structure for the home page
type HomeData struct {
	IsLoggedIn bool
	User       *auth.User
	Posts      []posts.Post
}

// serveHome handles requests to "/"
func ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := HomeData{} // Use the named struct instead of anonymous struct

	// Check if user is logged in
	userID, isLoggedIn := auth.GetUserIDFromSession(r)
	data.IsLoggedIn = isLoggedIn

	if isLoggedIn {
		// Get user data
		user, err := auth.GetUserByID(database, userID)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
		} else {
			data.User = user
		}
	}

	// Get posts
	postService := posts.NewPostService(database)
	allPosts, err := postService.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
		data.Posts = make([]posts.Post, 0) // Initialize empty slice
	} else {
		data.Posts = allPosts
	}

	// Set headers before writing any response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute template only once
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		if !isResponseWritten(w) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
}

// middleware for logging requests
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		handler(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.Println("Starting the Forum App...")

	// Initialize templates
	var err error
	templatePath := filepath.Join("internal", "web", "templates", "*.html")
	templates, err = template.ParseGlob(templatePath)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}

	// Ensure the correct number of arguments
	if len(os.Args) != 1 {
		fmt.Println("Invalid number of arguments.")
		fmt.Println("Usage: go run ./cmd/main.go")
		return
	}

	// Initialize the database
	database, err = db.InitializeDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize services
	postService := posts.NewPostService(database)
	postHandler := posts.NewPostHandler(postService, templates)
	authHandler := auth.NewAuthHandler(database, templates)
	commentHandler := comments.NewCommentHandler(database)
	likeService := likes.NewLikeService(database)
	likeHandler := likes.NewLikeHandler(likeService)

	// Register routes
	http.HandleFunc("/", logRequest(ServeHome))
	http.HandleFunc("/login", logRequest(authHandler.LoginHandler))
	http.HandleFunc("/register", logRequest(authHandler.RegisterHandler))
	http.HandleFunc("/logout", logRequest(authHandler.LogoutHandler))
	http.HandleFunc("/create", logRequest(postHandler.CreatePostHandler))
	http.HandleFunc("/posts", logRequest(postHandler.GetAllPostsHandler))
	http.HandleFunc("/profile/", logRequest(authHandler.ProfileHandler))
	http.HandleFunc("/comments", commentHandler.CreateComment)
	http.HandleFunc("/comments/post", commentHandler.GetCommentsForPost)
	http.HandleFunc("/react", likeHandler.HandleLike)

	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("internal/web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Create uploads directory
	if err := os.MkdirAll("internal/web/static/uploads", 0o755); err != nil {
		log.Printf("Warning: Failed to create uploads directory: %v", err)
	}

	// Start the server
	log.Printf("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// Helper function to check if response has been written
func isResponseWritten(w http.ResponseWriter) bool {
	rw, ok := w.(interface {
		Written() bool
	})
	return ok && rw.Written()
}
