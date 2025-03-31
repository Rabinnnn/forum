package testFiles

import (
    "testing"
    "golang.org/x/crypto/bcrypt"
    "forum/internal/auth"
)

func TestHashPassword(t *testing.T) {
    password := "securepassword"
    hash, err := auth.HashPassword(password)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if len(hash) == 0 {
        t.Fatal("hash should not be empty")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
        t.Fatal("hashed password does not match original password")
    }
}

func TestCheckPassword(t *testing.T) {
    password := "securepassword"
    hash, err := auth.HashPassword(password)
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if !auth.CheckPassword(password, hash) {
        t.Fatal("CheckPassword should return true for correct password")
    }
    
    if auth.CheckPassword("wrongpassword", hash) {
        t.Fatal("CheckPassword should return false for incorrect password")
    }
}
