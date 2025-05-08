package auth

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// secureStaticHandler prevents directory listing and restricts access to specific file types
func SecureStaticHandler() http.Handler {
	fs := http.FileServer(http.Dir("internal/web/static"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block directory listing attempts
		if strings.HasSuffix(r.URL.Path, "/") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Check file extension to allow only specific file types
		ext := filepath.Ext(r.URL.Path)
		allowedExts := map[string]bool{
			".css":   true,
			".js":    true,
			".png":   true,
			".jpg":   true,
			".jpeg":  true,
			".gif":   true,
			".svg":   true,
			".ico":   true,
			".woff":  true,
			".woff2": true,
			".ttf":   true,
		}

		if ext == "" || !allowedExts[ext] {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		log.Printf("Static file access: %s", r.URL.Path)

		// Serve the file if all checks pass
		fs.ServeHTTP(w, r)
	})
}

// secureUploadHandler provides secure access to uploaded files
func SecureUploadHandler() http.Handler {
	fs := http.FileServer(http.Dir("internal/web/static/uploads"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify user is authenticated before allowing access to uploads
		userID, isLoggedIn := GetUserIDFromSession(r)
		if !isLoggedIn {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Block directory listing attempts
		if strings.HasSuffix(r.URL.Path, "/") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Log access attempts
		log.Printf("Upload access by user %T: %s", userID, r.URL.Path)

		// Serve the file if all checks pass
		fs.ServeHTTP(w, r)
	})
}