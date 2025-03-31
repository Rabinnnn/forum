package testFiles

import (
    "testing"
    "time"
    "forum/internal/auth" 
)

func TestGenerateAndVerify(t *testing.T) {
    token := auth.Generate()
    if token == "" {
        t.Fatal("Generated token should not be empty")
    }

    if !auth.Verify(token) {
        t.Fatal("Verify should return true for a valid token")
    }

    if auth.Verify(token) {
        t.Fatal("Verify should return false for a used token")
    }
}

func TestVerifyExpiredToken(t *testing.T) {
    token := auth.Generate()
    auth.Mutex.Lock()
    auth.Tokens[token] = auth.CsrfToken{
        Token:   token,
        Expires: time.Now().Add(-1 * time.Hour), // Expired token
    }
    auth.Mutex.Unlock()

    if auth.Verify(token) {
        t.Fatal("Verify should return false for an expired token")
    }
}

