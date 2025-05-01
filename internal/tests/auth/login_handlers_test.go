package auth

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockDBLogin struct {
	queryRowFunc func(query string, args ...interface{}) *sql.Row
}

func (m *mockDBLogin) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.queryRowFunc(query, args...)
}

type DBInterface interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}

type TestAuthHandler struct {
	db        DBInterface
	templates TemplateInterface
}

// LoginHandler handles login requests
func (h *TestAuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		row := h.db.QueryRow("SELECT id, username, email, password FROM users WHERE username = ?", username)
		var id int
		var dbUsername, email, hashedPassword string
		if err := row.Scan(&id, &dbUsername, &email, &hashedPassword); err != nil {
			h.templates.ExecuteTemplate(w, "login.html", "Invalid credentials")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			h.templates.ExecuteTemplate(w, "login.html", "Invalid credentials")
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}


type TemplateInterface interface {
	ExecuteTemplate(wr http.ResponseWriter, name string, data interface{}) error
}

type mockTemplate struct {
	executeTemplateFunc func(wr http.ResponseWriter, name string, data interface{}) error
}

func (m *mockTemplate) ExecuteTemplate(wr http.ResponseWriter, name string, data interface{}) error {
	return m.executeTemplateFunc(wr, name, data)
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		formData       string
		dbQueryRowFunc func(query string, args ...interface{}) *sql.Row
		templateFunc   func(wr http.ResponseWriter, name string, data interface{}) error
		expectedCode   int
		expectedBody   string
	}{
		{
			name:   "GET request renders login page",
			method: http.MethodGet,
			templateFunc: func(wr http.ResponseWriter, name string, data interface{}) error {
				if name != "login.html" {
					t.Errorf("expected template name 'login.html', got '%s'", name)
				}
				return nil
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "POST request with invalid credentials",
			method:   http.MethodPost,
			formData: "username=testuser&password=wrongpassword",
			dbQueryRowFunc: func(query string, args ...interface{}) *sql.Row {
				return &sql.Row{}
			},
			templateFunc: func(wr http.ResponseWriter, name string, data interface{}) error {
				if name != "login.html" {
					t.Errorf("expected template name 'login.html', got '%s'", name)
				}
				return nil
			},
			expectedCode: http.StatusOK,
		},
		{
			name:     "POST request with valid credentials",
			method:   http.MethodPost,
			formData: "username=testuser&password=correctpassword",
			dbQueryRowFunc: func(query string, args ...interface{}) *sql.Row {
				hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
				return createMockRow([]interface{}{1, "testuser", "test@example.com", string(hashedPassword)})
			},
			templateFunc: func(wr http.ResponseWriter, name string, data interface{}) error {
				return nil
			},
			expectedCode: http.StatusSeeOther,
		},
		{
			name:         "Unsupported HTTP method",
			method:       http.MethodPut,
			expectedCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &mockDBLogin{queryRowFunc: tt.dbQueryRowFunc}
			templates := &mockTemplate{executeTemplateFunc: tt.templateFunc}
			handler := &TestAuthHandler{db: db, templates: templates}

			var body *bytes.Buffer
			if tt.formData != "" {
				body = bytes.NewBufferString(tt.formData)
			} else {
				body = &bytes.Buffer{}
			}

			req := httptest.NewRequest(tt.method, "/login", body)
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.LoginHandler(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, resp.StatusCode)
			}

			if tt.expectedBody != "" {
				body := w.Body.String()
				if !strings.Contains(body, tt.expectedBody) {
					t.Errorf("expected response body to contain '%s', got '%s'", tt.expectedBody, body)
				}
			}
		})
	}
}

func createMockRow(values []interface{}) *sql.Row {
	row := sql.Row{}
	row.Scan(values...)
	return &row
}
