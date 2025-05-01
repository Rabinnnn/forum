package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test CreateSession sets a session cookie and stores the user ID
func TestCreateSession(t *testing.T) {
	rr := httptest.NewRecorder()
	userID := "test-user"

	CreateSession(rr, userID)

	// Check that a cookie is set
	cookies := rr.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("Expected a session cookie to be set")
	}

	sessionCookie := cookies[0]
	if sessionCookie.Name != "session_id" {
		t.Errorf("Expected cookie name 'session_id', got %s", sessionCookie.Name)
	}

	// Check that the session ID maps to the correct userID
	if storedUserID, exists := sessions[sessionCookie.Value]; !exists || storedUserID != userID {
		t.Errorf("Session not stored correctly: got userID=%s, exists=%v", storedUserID, exists)
	}
}

// Test GetUserIDFromSession retrieves user ID from session cookie
func TestGetUserIDFromSession(t *testing.T) {
	// Create a fake session
	sessionID := "fake-session-id"
	userID := "user-123"
	sessions[sessionID] = userID

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})

	gotUserID, ok := GetUserIDFromSession(req)
	if !ok {
		t.Fatal("Expected session to exist but got false")
	}
	if gotUserID != userID {
		t.Errorf("Expected userID=%s, got %s", userID, gotUserID)
	}
}

// Test GetUserIDFromSession returns false if cookie is missing
func TestGetUserIDFromSession_MissingCookie(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	_, ok := GetUserIDFromSession(req)
	if ok {
		t.Error("Expected false when no session cookie exists")
	}
}

// Test ClearSession deletes session and clears cookie
func TestClearSession(t *testing.T) {
	// Create session
	sessionID := "clear-session-id"
	userID := "user-456"
	sessions[sessionID] = userID

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	rr := httptest.NewRecorder()

	ClearSession(rr, req)

	// Check session is removed
	if _, exists := sessions[sessionID]; exists {
		t.Error("Expected session to be deleted")
	}

	// Check that cookie is cleared in response
	found := false
	for _, cookie := range rr.Result().Cookies() {
		if cookie.Name == "session_id" && cookie.Value == "" && cookie.MaxAge == -1 {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected session cookie to be cleared with MaxAge -1")
	}
}
