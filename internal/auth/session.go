package auth

import (
	"log"
	"net/http"
	"github.com/google/uuid"
)

var sessions = make(map[string]string) // sessionID -> userID

func CreateSession(w http.ResponseWriter, userID string) {
	sessionID := uuid.New().String()
	sessions[sessionID] = userID

	log.Printf("Creating new session - SessionID: %s, UserID: %s", sessionID, userID)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		// fields to make the cookie more secure and persistent
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400 * 30, // 30 days
	}
	http.SetCookie(w, cookie)
}

func GetUserIDFromSession(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		return "", false
	}
	
	sessionID := cookie.Value
	userID, exists := sessions[sessionID]
	
	log.Printf("Session check - SessionID: %s, UserID: %s, Exists: %v", sessionID, userID, exists)
	
	return userID, exists
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		delete(sessions, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
