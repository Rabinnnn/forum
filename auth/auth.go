package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

// GitHub OAuth URLs
const (
	authURL  = "https://github.com/login/oauth/authorize"
	tokenURL = "https://github.com/login/oauth/access_token"
	userURL  = "https://api.github.com/user"
)

// Login redirects the user to GitHub OAuth page
func Login(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	redirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	oauthURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&scope=read:user", authURL, clientID, redirectURL)
	http.Redirect(w, r, oauthURL, http.StatusFound)
}

// Callback handles OAuth response from GitHub
func Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code in callback", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	accessToken, err := getAccessToken(code)
	if err != nil {
		http.Error(w, "Failed to get access token", http.StatusInternalServerError)
		return
	}

	// Fetch user info from GitHub
	userData, err := getUserData(accessToken)
	if err != nil {
		http.Error(w, "Failed to fetch user data", http.StatusInternalServerError)
		return
	}

	// Display user data
	fmt.Fprintf(w, "Logged in as: %s", userData.Login)
}

// Exchange code for access token
func getAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	data.Set("code", code)

	req, _ := http.NewRequest("POST", tokenURL, nil)
	req.URL.RawQuery = data.Encode()
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var res struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(resp.Body).Decode(&res)

	return res.AccessToken, nil
}

// Fetch user info from GitHub API
func getUserData(token string) (*GitHubUser, error) {
	req, _ := http.NewRequest("GET", userURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var user GitHubUser
	json.Unmarshal(body, &user)

	return &user, nil
}

// GitHub user data structure
type GitHubUser struct {
	Login string `json:"login"`
}
