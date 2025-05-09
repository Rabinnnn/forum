package auth

import (
	"bytes"
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginHandler_Get(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	h := &AuthHandler{
		db: db,
		templates: template.Must(template.New("login.html").Parse(`
			<html><body>Login Page</body></html>
		`)),
	}

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestLoginHandler_Post_Success(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	password := "secret"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, username, email, password FROM users").
		WithArgs("user", "user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password"}).
			AddRow("123", "user", "user@example.com", string(hashedPassword)))

	h := &AuthHandler{
		db: db,
		templates: template.Must(template.New("login.html").Parse(`
			<html><body>Login Page</body></html>
		`)),
	}

	form := "username=user&password=secret"
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected status 303 SeeOther, got %d", resp.StatusCode)
	}
}

func TestLoginHandler_Post_InvalidPassword(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)

	mock.ExpectQuery("SELECT id, username, email, password FROM users").
		WithArgs("user", "user").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "password"}).
			AddRow("123", "user", "user@example.com", string(hashedPassword)))

	h := &AuthHandler{
		db: db,
		templates: template.Must(template.New("login.html").Parse(`
			{{if .Error}}<p>{{.Error}}</p>{{end}}
		`)),
	}

	form := "username=user&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 with error message, got %d", resp.StatusCode)
	}
}

func TestLoginHandler_Post_UserNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT id, username, email, password FROM users").
		WithArgs("ghost", "ghost").
		WillReturnError(sql.ErrNoRows)

	h := &AuthHandler{
		db: db,
		templates: template.Must(template.New("login.html").Parse(`
			{{if .Error}}<p>{{.Error}}</p>{{end}}
		`)),
	}

	form := "username=ghost&password=anything"
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 with error message, got %d", resp.StatusCode)
	}
}

func TestLoginHandler_InvalidMethod(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	h := &AuthHandler{
		db: db,
		templates: template.Must(template.New("login.html").Parse(`
			<html><body>Login Page</body></html>
		`)),
	}

	req := httptest.NewRequest(http.MethodPut, "/login", nil)
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", resp.StatusCode)
	}
}
