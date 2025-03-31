package testFiles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"forum/internal/auth"
)

func TestCreateSession(t *testing.T) {
	w := httptest.NewRecorder()
	userID := "test_user"

	auth.CreateSession(w, userID)

	// Check if the cookie is set
	resp := w.Result()
	cookie := resp.Cookies()

	if len(cookie) == 0 || cookie[0].Name != "session_id" {
		t.Fatal("Session cookie was not set")
	}

	sessionID := cookie[0].Value
	if storedUser, exists := auth.Sessions[sessionID]; !exists || storedUser != userID {
		t.Fatalf("Session ID not stored correctly. Got: %s, Expected: %s", storedUser, userID)
	}
}

func TestGetUserIDFromSession(t *testing.T) {
	// Simulate an existing session
	sessionID := "test_session"
	userID := "test_user"
	auth.Sessions[sessionID] = userID

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

	retrievedUserID, exists := auth.GetUserIDFromSession(req)

	if !exists {
		t.Fatal("Session lookup failed when it should have succeeded")
	}
	if retrievedUserID != userID {
		t.Fatalf("Expected userID %s, got %s", userID, retrievedUserID)
	}
}

func TestClearSession(t *testing.T) {
	// Simulate an existing session
	sessionID := "test_session"
	userID := "test_user"
	auth.Sessions[sessionID] = userID

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

	w := httptest.NewRecorder()
	auth.ClearSession(w, req)

	// Check if the session was removed
	if _, exists := auth.Sessions[sessionID]; exists {
		t.Fatal("Session was not cleared")
	}

	// Check if the cookie was expired
	resp := w.Result()
	cookies := resp.Cookies()

	if len(cookies) == 0 || cookies[0].MaxAge != -1 {
		t.Fatal("Session cookie was not expired")
	}
}