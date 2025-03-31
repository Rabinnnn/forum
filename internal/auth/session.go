package auth

import (
	"log"
	"net/http"
	"github.com/google/uuid"
)

var Sessions = make(map[string]string) // sessionID -> userID

func CreateSession(w http.ResponseWriter, userID string) {
	sessionID := uuid.New().String()
	Sessions[sessionID] = userID

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

	userID, exists := Sessions[cookie.Value]
	log.Printf("Session lookup: cookie=%s, userID=%s, exists=%v",
		cookie.Value, userID, exists)

	return userID, exists
}

func ClearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		delete(Sessions, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
}
