package auth

import (
	"testing"
	"time"
)

func TestGenerateAndVerify(t *testing.T) {
	token := Generate()
	if token == "" {
		t.Fatal("Expected non-empty token from Generate")
	}

	if !Verify(token) {
		t.Error("Expected token to be valid immediately after generation")
	}

	// Token should not be reusable
	if Verify(token) {
		t.Error("Expected token to be invalid after being used once")
	}
}

func TestExpiredToken(t *testing.T) {
	token := Generate()

	// Simulate token expiration
	mutex.Lock()
	tokens[token] = csrfToken{
		token:   token,
		expires: time.Now().Add(-1 * time.Minute), // expired
	}
	mutex.Unlock()

	if Verify(token) {
		t.Error("Expected expired token to be invalid")
	}
}
