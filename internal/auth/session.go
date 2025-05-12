package auth

import (
	"github.com/google/uuid"
	"log"
	"net/http"
)

var sessions = make(map[string]string) // sessionID -> userID

func CreateSession(w http.ResponseWriter, userID string) {
	for sessionID, id := range sessions {
		if id == userID {
			delete(sessions, sessionID)
			log.Printf("Deleted existing session: ID=%s for user=%s", sessionID, userID)
			break // Assuming one session per user; remove `break` if multiple could exist
		}
	}

	sessionID := uuid.New().String()
	sessions[sessionID] = userID

	log.Printf("Creating session: ID=%s for user=%s", sessionID, userID)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func GetUserIDFromSession(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("No session cookie found: %v", err)
		return "", false
	}

	userID, exists := sessions[cookie.Value]
	log.Printf("Session lookup: cookie=%s, userID=%s, exists=%v",
		cookie.Value, userID, exists)

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
