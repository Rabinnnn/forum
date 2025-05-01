package auth

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
    password := "mySecureP@ssw0rd"

    hashed, err := HashPassword(password)
    if err != nil {
        t.Fatalf("Error hashing password: %v", err)
    }

    if hashed == password {
        t.Error("Hashed password should not be the same as the original password")
    }

    if !CheckPassword(password, hashed) {
        t.Error("CheckPassword failed: password should match the hash")
    }

    if CheckPassword("wrongPassword", hashed) {
        t.Error("CheckPassword should fail with wrong password")
    }
}
