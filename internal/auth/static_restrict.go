package auth

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"forum/internal/xerrors"
)

func SecureStaticHandler() http.Handler {
	fs := http.FileServer(http.Dir("internal/web/static"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block directory listing
		if strings.HasSuffix(r.URL.Path, "/") {
			xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrForbidden)
			return
		}

		// Allow only specific extensions
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
			xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrForbidden)
			return
		}

		// Reject direct navigation (e.g., user typing URL in browser)
		accept := r.Header.Get("Accept")
		if strings.Contains(accept, "text/html") {
			xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrForbidden)
			return
		}

		log.Printf("Static file access: %s", r.URL.Path)
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
			xerrors.RenderErrorPage(w, http.StatusUnauthorized, xerrors.ErrUnauthorized)
			return
		}

		// Block directory listing attempts
		if strings.HasSuffix(r.URL.Path, "/") {
			xerrors.RenderErrorPage(w, http.StatusForbidden, xerrors.ErrForbidden)
			return
		}

		// Log access attempts
		log.Printf("Upload access by user %T: %s", userID, r.URL.Path)

		// Serve the file if all checks pass
		fs.ServeHTTP(w, r)
	})
}
