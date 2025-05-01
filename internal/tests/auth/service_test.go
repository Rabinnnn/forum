package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"html/template"

	"github.com/DATA-DOG/go-sqlmock"
)




func TestGetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error mocking db: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "email", "profile_pic"}).
		AddRow("123", "testuser", "test@example.com", "pic.jpg")

	mock.ExpectQuery("SELECT id, username, email, COALESCE.*").
		WithArgs("123").
		WillReturnRows(rows)

	user, err := GetUserByID(db, "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil || user.ID != "123" {
		t.Fatalf("expected user ID 123, got %+v", user)
	}
}

func TestLogoutHandler(t *testing.T) {
	handler := &AuthHandler{}

	req := httptest.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()

	handler.LogoutHandler(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("expected status %v, got %v", http.StatusSeeOther, status)
	}
	if loc := rr.Header().Get("Location"); loc != "/" {
		t.Errorf("expected redirect to '/', got %s", loc)
	}
}

func TestProfileHandler_NotFound(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	tmpl := template.Must(template.New("profile.html").Parse(`
		<html><body>
		User: {{.Username}}<br>
		Email: {{.Email}}<br>
		{{range .Posts}}<div>{{.Title}}</div>{{end}}
		</body></html>
	`))
	handler := NewAuthHandler(db, tmpl)

	req := httptest.NewRequest("GET", "/profile/", nil)
	rr := httptest.NewRecorder()

	handler.ProfileHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestProfileHandler_RenderSuccess(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	// mock user
	rows := sqlmock.NewRows([]string{"id", "username", "email", "profile_pic"}).
		AddRow("123", "granton", "test@example.com", "pic.jpg")
	mock.ExpectQuery("SELECT id, username, email, COALESCE.*").
		WithArgs("123").
		WillReturnRows(rows)

	// mock posts
	postRows := sqlmock.NewRows([]string{
		"id", "title", "content", "image_path", "created_at", "likes", "dislikes", "comments",
	}).AddRow(1, "Post Title", "Post content", "img.png", "2024-01-01", 3, 1, 2)

	mock.ExpectQuery("SELECT \\s+p.id, p.title.*").
		WithArgs("123").
		WillReturnRows(postRows)

	tmpl := template.Must(template.New("profile.html").Parse(`
		<html><body>
		User: {{.Username}}<br>
		Email: {{.Email}}<br>
		{{range .Posts}}<div>{{.Title}}</div>{{end}}
		</body></html>
	`))	
	handler := NewAuthHandler(db, tmpl)

	req := httptest.NewRequest("GET", "/profile/123", nil)
	rr := httptest.NewRecorder()

	handler.ProfileHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "granton") {
		t.Errorf("expected username in response, got %s", rr.Body.String())
	}
}
