package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "Valid Password",
			password: "securepassword123",
		},
		{
			name:     "Empty Password",
			password: "",
		},
		{
			name:     "Special Characters",
			password: "!@#$%^&*()_+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if hash == "" {
				t.Errorf("expected non-empty hash, got empty string")
			}

			// Verify the hash matches the original password
			if !CheckPassword(tt.password, hash) {
				t.Errorf("password does not match hash")
			}
		})
	}
}
