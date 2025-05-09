package auth

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "mySecurePassword123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned an error: %v", err)
	}

	if hashedPassword == "" {
		t.Fatal("HashPassword returned an empty string")
	}

	if !CheckPassword(password, hashedPassword) {
		t.Error("CheckPassword failed: correct password was not recognized")
	}

	wrongPassword := "wrongPassword"
	if CheckPassword(wrongPassword, hashedPassword) {
		t.Error("CheckPassword incorrectly returned true for wrong password")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "repeatPassword"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error hashing password: %v, %v", err1, err2)
	}

	if hash1 == hash2 {
		t.Error("Expected different hashes for same password due to salt, but got identical hashes")
	}
}
