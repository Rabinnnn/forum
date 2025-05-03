package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"forum/internal/likes"
	"forum/internal/filters"
	"forum/internal/comments"
	"forum/internal/auth"
	"forum/internal/db"
	"forum/internal/posts"
)

// Global variables
var (
	templates *template.Template
	database  *sql.DB
)

// serveHome handles requests to "/"
func ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		IsLoggedIn bool
		User       *auth.User
		Posts      []posts.Post
	}{}

	// Check if user is logged in
	userID, isLoggedIn := auth.GetUserIDFromSession(r)
	data.IsLoggedIn = isLoggedIn

	if isLoggedIn {
		// Get user data
		user, err := auth.GetUserByID(database, userID)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
		}else if user == nil {
			// Handle "not found" case
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else {
			data.User = user
		}
	}

	// Get posts
	postService := posts.NewPostService(database)
	posts, err := postService.GetAllPosts()
	if err != nil {
		log.Printf("Error fetching posts: %v", err)
	} else {
		data.Posts = posts
	}

	log.Printf("Template data: IsLoggedIn=%v, User=%+v", data.IsLoggedIn, data.User)

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
	
}

// middleware for logging requests
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.Path)
		log.Printf("Cookies: %v", r.Cookies())
		handler(w, r)
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


	// Register routes
	http.HandleFunc("/", logRequest(ServeHome))
	http.HandleFunc("/login", logRequest(authHandler.LoginHandler))
	http.HandleFunc("/register", logRequest(authHandler.RegisterHandler))
	http.HandleFunc("/logout", logRequest(authHandler.LogoutHandler))
	http.HandleFunc("/create", logRequest(postHandler.CreatePostHandler))
	http.HandleFunc("/posts", logRequest(postHandler.GetAllPostsHandler))
	http.HandleFunc("/profile/", logRequest(authHandler.ProfileHandler)) // Note the trailing slash
	http.HandleFunc("/comments", commentHandler.CreateComment)
	http.HandleFunc("/comments/post", commentHandler.GetCommentsForPost)
	http.HandleFunc("/editcomment", comments.HandleEditComment)
	http.HandleFunc("/deletecomment", comments.HandleDeleteComment)
	http.HandleFunc("/commentreact", comments.HandleCommentReactions)



	http.HandleFunc("/react", likes.HandleReactions)
	http.HandleFunc("/created", filters.CreatedPosts)
	categoryHandler := filters.NewCategoryHandler()
	http.Handle("/categories", categoryHandler)
	http.Handle("/category", categoryHandler)




	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("internal/web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Create uploads directory
	if err := os.MkdirAll("internal/web/static/uploads", 0o755); err != nil {
		log.Printf("Warning: Failed to create uploads directory: %v", err)
	}

	// Start the server
	log.Printf("Server starting on http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
