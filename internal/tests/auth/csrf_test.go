package auth

import (
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	token := Generate()

	if token == "" {
		t.Fatal("Generated token should not be empty")
	}

	// Check if the token is stored with an expiration
	mutex.RLock()
	storedToken, exists := tokens[token]
	mutex.RUnlock()

	if !exists {
		t.Fatal("Generated token should be stored in the tokens map")
	}

	if time.Now().After(storedToken.expires) {
		t.Fatal("Generated token should have a valid expiration time")
	}

	if storedToken.token != token {
		t.Fatal("Stored token should match the generated token")
	}
}
