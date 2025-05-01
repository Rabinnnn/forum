package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"html/template"
	"strings"
)

func TestRegisterHandler(t *testing.T) {
	// Create a mock AuthHandler with templates and a mock DB connection
	handler := &AuthHandler{
		templates: template.Must(template.ParseGlob("test_templates/*.html")),
		// Mock your DB here if necessary
	}

	// Test: Get request to load the registration form
	req, err := http.NewRequest(http.MethodGet, "/register", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	handler.RegisterHandler(rr, req)

	// Check if the status code is correct
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Test: Post request with invalid input (empty fields)
	formData := "username=&email=&password=&confirm-password="
	req, err = http.NewRequest(http.MethodPost, "/register", strings.NewReader(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.RegisterHandler(rr, req)

	// Check if validation errors are returned
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	// Test: Post request with valid input
	formData = "username=johndoe&email=johndoe@example.com&password=securepassword&confirm-password=securepassword"
	req, err = http.NewRequest(http.MethodPost, "/register", strings.NewReader(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.RegisterHandler(rr, req)

	// Check if redirect occurs after successful registration
	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, status)
	}
}

func TestRegisterHandlerWithExistingUser(t *testing.T) {
	// Mock an existing user scenario
	handler := &AuthHandler{
		templates: template.Must(template.ParseGlob("test_templates/*.html")),
		// Mock your DB connection
	}

	// Test: Post request with existing username/email
	formData := "username=johndoe&email=johndoe@example.com&password=securepassword&confirm-password=securepassword"
	req, err := http.NewRequest(http.MethodPost, "/register", strings.NewReader(formData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.RegisterHandler(rr, req)

	// Check if error is returned for existing username/email
	if rr.Body.String() != "Username or email already exists" {
		t.Errorf("Expected error for existing user, got: %s", rr.Body.String())
	}
}