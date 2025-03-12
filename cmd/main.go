package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"forum/internal/db"
)

// serveHome handles requests to "/"
func serveHome(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/web/templates/index.html") // Updated path
}

func main() {
	fmt.Println("Starting the Forum App...")

	// Ensure the correct number of arguments
	if len(os.Args) != 1 {
		fmt.Println("Invalid number of arguments.")
		fmt.Println("Usage: go run ./cmd/main.go")
		return
	}

	// Initialize the database
	database, err := db.InitializeDB()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}
	defer database.Close()

	// Define routes
	http.HandleFunc("/", serveHome)

	// Serve static files (CSS, JS, images)
	fs := http.FileServer(http.Dir("internal/web/static")) // Updated path
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	port := ":8080"
	fmt.Println("Server started at http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}
